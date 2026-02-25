package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdp/qrterminal/v3"
	"github.com/spf13/cast"
	"golang.org/x/oauth2"
	"schneider.vip/share/clipboard"
	"schneider.vip/share/config"
	"schneider.vip/share/provider"
	"schneider.vip/share/provider/box"
	"schneider.vip/share/provider/dropbox"
	"schneider.vip/share/provider/googledrive"
	"schneider.vip/share/provider/httpupload"
	"schneider.vip/share/provider/nextcloud"
	"schneider.vip/share/provider/opendrive"
	"schneider.vip/share/provider/seafile"
	"schneider.vip/share/tui/setup"
	"schneider.vip/share/tui/upload"
)

var version = "0.0.0"

type CLI struct {
	Config  string   `help:"Path to config file (default: ${defaultConfigPath})." type:"path"`
	Setup   bool     `help:"Launch interactive setup." short:"s"`
	Select  bool     `help:"Select provider for this upload." short:"p"`
	Version bool     `help:"Print version and exit." short:"v"`
	Args    []string `arg:"" optional:"" help:"File to upload and optional provider name."`
}

func main() {
	cli := CLI{}
	kong.Parse(&cli,
		kong.Name("share"),
		kong.Description("Upload files to cloud storage and get a shareable link."),
		kong.UsageOnError(),
		kong.Vars{
			"defaultConfigPath": config.DefaultConfigPath(),
		},
	)

	if cli.Version {
		fmt.Printf("ShareCmd Version: %s\n", version)
		os.Exit(0)
	}

	configPath := cli.Config
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	cfg, err := config.LookupConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	if cli.Setup || cfg.ActiveProvider() == nil {
		if err := setup.Run(cfg); err != nil {
			log.Fatalf("Setup failed: %v\n", err)
		}
		if len(cli.Args) == 0 {
			os.Exit(0)
		}
		// Reload config after setup
		cfg, err = config.LookupConfig(configPath)
		if err != nil {
			log.Fatalf("Failed to reload config: %v\n", err)
		}
	}

	// Parse args: extract filename and optional provider override
	if len(cli.Args) == 0 {
		os.Exit(0)
	}

	if len(cli.Args) > 2 {
		log.Fatalf("Too many arguments. Usage: share [provider] <file> or share <file> [provider]\n")
	}

	var filename string
	var providerLabel string

	for _, arg := range cli.Args {
		// First check if it's an existing file (to handle files named like providers)
		if _, err := os.Stat(arg); err == nil {
			if filename != "" {
				log.Fatalf("Multiple files specified: %q and %q\n", filename, arg)
			}
			filename = arg
		} else if cfg.FindByLabel(arg) != nil {
			// It's a provider label
			if providerLabel != "" {
				log.Fatalf("Multiple providers specified: %q and %q\n", providerLabel, arg)
			}
			providerLabel = arg
		} else {
			// Neither file nor provider - show interactive selector
			if len(cfg.Providers) == 0 {
				log.Fatalf("Argument %q is neither an existing file nor a configured provider. No providers configured.\n", arg)
			}
			fmt.Printf("Argument %q is not a configured provider.\n", arg)
			selected, err := setup.PickProvider(cfg, "Select provider for this upload")
			if err != nil {
				log.Fatalf("Provider selection cancelled: %v\n", err)
			}
			if providerLabel != "" {
				log.Fatalf("Multiple providers specified: %q and %q\n", providerLabel, selected)
			}
			providerLabel = selected
		}
	}

	if filename == "" {
		log.Fatalf("No file to upload specified\n")
	}

	// Determine which provider to use
	var active *config.ProviderEntry
	if cli.Select || providerLabel != "" {
		// Interactive provider selection
		if providerLabel == "" {
			if len(cfg.Providers) == 0 {
				log.Fatal("No providers configured. Run 'share --setup' to configure.")
			}
			selected, err := setup.PickProvider(cfg, "Select provider for this upload")
			if err != nil {
				log.Fatalf("Provider selection cancelled: %v\n", err)
			}
			providerLabel = selected
		}
		active = cfg.FindByLabel(providerLabel)
	} else {
		active = cfg.ActiveProvider()
		if active == nil {
			log.Fatal("No active provider configured. Run 'share --setup' to configure.")
		}
	}

	prov, err := instantiateProvider(active)
	if err != nil {
		log.Fatalf("Failed to create provider: %v\n", err)
	}

	// Setup token refresh callback for OAuth2 providers
	setupTokenRefresh(prov, active, cfg)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Can't open file: %v\n", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Can't stat file: %v\n", err)
	}
	basename := filepath.Base(file.Name())
	filesize := fileInfo.Size()

	// Upload with progress TUI
	var fileID string
	var uploadErr error

	model := upload.NewModel(basename, filesize)
	p := tea.NewProgram(model)
	pr := upload.NewProgressReader(file, filesize, p)

	go func() {
		fileID, uploadErr = prov.Upload(pr, basename, filesize)
		upload.SendDone(p, fileID, uploadErr)
	}()

	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI error: %v\n", err)
	}

	if uploadErr != nil {
		if isOAuthTokenError(uploadErr) {
			fmt.Printf("\nOAuth token has expired for provider %q.\n", active.Label)
			if err := setup.ReconfigureProvider(cfg, active.Label); err != nil {
				log.Fatalf("Re-authentication failed: %v\n", err)
			}
			// Reload config and retry
			cfg, err = config.LookupConfig(configPath)
			if err != nil {
				log.Fatalf("Failed to reload config: %v\n", err)
			}
			active = cfg.FindByLabel(active.Label)
			prov, err = instantiateProvider(active)
			if err != nil {
				log.Fatalf("Failed to create provider: %v\n", err)
			}
			setupTokenRefresh(prov, active, cfg)

			// Retry upload
			file.Seek(0, 0)
			pr2 := upload.NewProgressReader(file, filesize, p)
			fileID, uploadErr = prov.Upload(pr2, basename, filesize)
			if uploadErr != nil {
				log.Fatalf("Upload failed after re-authentication: %v\n", uploadErr)
			}
		} else {
			log.Fatalf("Upload failed: %v\n", uploadErr)
		}
	}

	link, err := prov.GetLink(fileID)
	if err != nil {
		if isOAuthTokenError(err) {
			fmt.Printf("\nOAuth token has expired for provider %q.\n", active.Label)
			if err := setup.ReconfigureProvider(cfg, active.Label); err != nil {
				log.Fatalf("Re-authentication failed: %v\n", err)
			}
			// Reload config and retry
			cfg, err = config.LookupConfig(configPath)
			if err != nil {
				log.Fatalf("Failed to reload config: %v\n", err)
			}
			active = cfg.FindByLabel(active.Label)
			prov, err = instantiateProvider(active)
			if err != nil {
				log.Fatalf("Failed to create provider: %v\n", err)
			}
			setupTokenRefresh(prov, active, cfg)

			// Retry GetLink
			link, err = prov.GetLink(fileID)
			if err != nil {
				log.Fatalf("GetLink failed after re-authentication: %v\n", err)
			}
		} else {
			log.Fatalf("Can't get link: %v\n", err)
		}
	}

	if cfg.ShowQRCodeEnabled() {
		fmt.Println()
		if qrterminal.IsSixelSupported(os.Stdout) {
			qrterminal.Generate(link, qrterminal.L, os.Stdout)
		} else {
			qrterminal.GenerateHalfBlock(link, qrterminal.L, os.Stdout)
		}
		fmt.Println()
	}
	fmt.Printf("URL: %s\n", link)

	if cfg.CopyToClipboardEnabled() {
		clipboard.ToClip(link)
	}
}

func instantiateProvider(entry *config.ProviderEntry) (provider.Provider, error) {
	switch entry.Type {
	case "httpupload":
		return httpupload.NewProvider(entry.Settings["url"], entry.Settings["headers"]), nil
	case "seafile":
		return seafile.NewProvider(entry.Settings["url"], entry.Settings["token"], entry.Settings["repoid"]), nil
	case "opendrive":
		return opendrive.NewProvider(entry.Settings["user"], entry.Settings["pass"]), nil
	case "dropbox":
		return dropbox.NewProvider(entry.Settings["token"]), nil
	case "nextcloud":
		return nextcloud.NewProvider(nextcloud.Config{
			URL:                   entry.Settings["url"],
			Username:              entry.Settings["username"],
			Password:              entry.Settings["password"],
			LinkShareWithPassword: cast.ToBool(entry.Settings["linkShareWithPassword"]),
			RandomPasswordChars:   cast.ToInt(entry.Settings["randomPasswordChars"]),
		}), nil
	case "box":
		return box.NewProvider(entry.Settings["token"]), nil
	case "googledrive":
		return googledrive.NewProvider(entry.Settings["googletoken"]), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", entry.Type)
	}
}

// OAuth2Provider is an interface for providers that support OAuth2 token refresh
type OAuth2Provider interface {
	SetTokenRefreshCallback(func(*oauth2.Token))
	GetCurrentToken() *oauth2.Token
}

func setupTokenRefresh(prov provider.Provider, entry *config.ProviderEntry, cfg *config.Config) {
	oauth2Prov, ok := prov.(OAuth2Provider)
	if !ok {
		return // Not an OAuth2 provider
	}

	oauth2Prov.SetTokenRefreshCallback(func(newToken *oauth2.Token) {
		tokenJSON, err := json.Marshal(newToken)
		if err != nil {
			log.Printf("Warning: failed to marshal refreshed token: %v\n", err)
			return
		}

		// Update the token in the config
		settingKey := "token"
		if entry.Type == "googledrive" {
			settingKey = "googletoken"
		}
		entry.Settings[settingKey] = string(tokenJSON)

		// Save config to disk
		if err := cfg.Write(); err != nil {
			log.Printf("Warning: failed to save refreshed token to config: %v\n", err)
		}
	})
}

// isOAuthTokenError checks if an error is related to expired/invalid OAuth tokens
func isOAuthTokenError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "invalid_grant") ||
		strings.Contains(errStr, "token has expired") ||
		strings.Contains(errStr, "Refresh token has expired") ||
		strings.Contains(errStr, "invalid_token") ||
		strings.Contains(errStr, "Token expired") ||
		strings.Contains(errStr, "unauthorized")
}
