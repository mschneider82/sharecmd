package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdp/qrterminal"
	"github.com/spf13/cast"
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
	Config  string   `help:"Path to config file." default:"~/.config/sharecmd/config.json" type:"path"`
	Setup   bool     `help:"Launch interactive setup." short:"s"`
	Version bool     `help:"Print version and exit." short:"v"`
	Args    []string `arg:"" optional:"" help:"File to upload and optional provider name."`
}

func main() {
	cli := CLI{}
	kong.Parse(&cli,
		kong.Name("share"),
		kong.Description("Upload files to cloud storage and get a shareable link."),
		kong.UsageOnError(),
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
	var filename string
	var providerLabel string

	for _, arg := range cli.Args {
		// Check if arg matches a configured provider label
		if cfg.FindByLabel(arg) != nil {
			providerLabel = arg
		} else {
			// Assume it's the filename
			filename = arg
		}
	}

	if filename == "" {
		os.Exit(0)
	}

	// Verify file exists
	if _, err := os.Stat(filename); err != nil {
		log.Fatalf("File not found: %s\n", filename)
	}

	// Determine which provider to use
	var active *config.ProviderEntry
	if providerLabel != "" {
		active = cfg.FindByLabel(providerLabel)
		if active == nil {
			log.Fatalf("Provider %q not found in config\n", providerLabel)
		}
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
		log.Fatalf("Upload failed: %v\n", uploadErr)
	}

	link, err := prov.GetLink(fileID)
	if err != nil {
		log.Fatalf("Can't get link: %v\n", err)
	}

	if cfg.ShowQRCodeEnabled() {
		var qr strings.Builder
		qrterminal.Generate(link, qrterminal.L, &qr)
		fmt.Printf("\n%s\n", qr.String())
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
