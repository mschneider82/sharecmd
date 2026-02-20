package setup

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/oauth2"
	"schneider.vip/share/config"
	"schneider.vip/share/provider/box"
	"schneider.vip/share/provider/dropbox"
	"schneider.vip/share/provider/googledrive"
	"schneider.vip/share/provider/seafile"
	"schneider.vip/share/tui"
)

// Run launches the interactive setup TUI. It loops a main menu until the user quits.
func Run(cfg *config.Config) error {
	for {
		action, err := mainMenu(cfg)
		if err != nil {
			return err
		}
		switch action {
		case "select":
			if err := selectActive(cfg); err != nil {
				return err
			}
			return nil
		case "add":
			if err := addProvider(cfg); err != nil {
				return err
			}
		case "edit":
			if err := editProvider(cfg); err != nil {
				return err
			}
		case "delete":
			if err := deleteProvider(cfg); err != nil {
				return err
			}
		case "preferences":
			if err := editPreferences(cfg); err != nil {
				return err
			}
		case "quit":
			return nil
		}
	}
}

func mainMenu(cfg *config.Config) (string, error) {
	var action string

	options := []huh.Option[string]{
		huh.NewOption("Add new provider", "add"),
	}
	if len(cfg.Providers) > 0 {
		options = append([]huh.Option[string]{
			huh.NewOption("Select active provider", "select"),
		}, options...)
		options = append(options,
			huh.NewOption("Edit provider", "edit"),
			huh.NewOption("Delete provider", "delete"),
		)
	}
	options = append(options,
		huh.NewOption("Preferences", "preferences"),
		huh.NewOption("Quit", "quit"),
	)

	title := tui.Title.Render("ShareCmd Setup")
	if cfg.Active != "" {
		title += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			fmt.Sprintf("  (active: %s)", cfg.Active))
	}
	fmt.Println(title)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(options...).
				Value(&action),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return action, nil
}

func selectActive(cfg *config.Config) error {
	if len(cfg.Providers) == 0 {
		fmt.Println(tui.Error.Render("No providers configured."))
		return nil
	}

	var label string
	options := make([]huh.Option[string], len(cfg.Providers))
	for i, p := range cfg.Providers {
		desc := p.Type
		if p.Label == cfg.Active {
			desc += " (current)"
		}
		options[i] = huh.NewOption(p.Label+" — "+desc, p.Label)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select active provider").
				Options(options...).
				Value(&label),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	cfg.SetActive(label)
	if err := cfg.Write(); err != nil {
		return err
	}
	fmt.Println(tui.Success.Render(fmt.Sprintf("Active provider set to %q", label)))
	return nil
}

func addProvider(cfg *config.Config) error {
	var provType string
	options := make([]huh.Option[string], len(ProviderTypes))
	for i, t := range ProviderTypes {
		options[i] = huh.NewOption(t, t)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select provider type").
				Options(options...).
				Value(&provType),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	settings, err := runProviderForm(provType, nil)
	if err != nil {
		return err
	}

	// Ask for a label
	lForm, label := labelForm(provType)
	if err := lForm.Run(); err != nil {
		return err
	}

	// Ensure unique label
	for cfg.FindByLabel(*label) != nil {
		fmt.Println(tui.Error.Render(fmt.Sprintf("Label %q already exists, choose another.", *label)))
		lForm, label = labelForm(*label + "-2")
		if err := lForm.Run(); err != nil {
			return err
		}
	}

	entry := config.ProviderEntry{
		Label:    *label,
		Type:     provType,
		Settings: settings,
	}
	cfg.AddProvider(entry)

	// If this is the only provider, set it as active.
	if len(cfg.Providers) == 1 {
		cfg.Active = entry.Label
	}

	if err := cfg.Write(); err != nil {
		return err
	}
	fmt.Println(tui.Success.Render(fmt.Sprintf("Provider %q added.", *label)))
	return nil
}

func editProvider(cfg *config.Config) error {
	if len(cfg.Providers) == 0 {
		fmt.Println(tui.Error.Render("No providers to edit."))
		return nil
	}

	label, err := pickProvider(cfg, "Select provider to edit")
	if err != nil {
		return err
	}

	entry := cfg.FindByLabel(label)
	if entry == nil {
		return fmt.Errorf("provider %q not found", label)
	}

	settings, err := runProviderForm(entry.Type, entry.Settings)
	if err != nil {
		return err
	}

	entry.Settings = settings
	if err := cfg.Write(); err != nil {
		return err
	}
	fmt.Println(tui.Success.Render(fmt.Sprintf("Provider %q updated.", label)))
	return nil
}

func deleteProvider(cfg *config.Config) error {
	if len(cfg.Providers) == 0 {
		fmt.Println(tui.Error.Render("No providers to delete."))
		return nil
	}

	label, err := pickProvider(cfg, "Select provider to delete")
	if err != nil {
		return err
	}

	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete provider %q?", label)).
				Value(&confirm),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	if !confirm {
		return nil
	}

	cfg.RemoveProvider(label)
	if err := cfg.Write(); err != nil {
		return err
	}
	fmt.Println(tui.Success.Render(fmt.Sprintf("Provider %q deleted.", label)))
	return nil
}

func editPreferences(cfg *config.Config) error {
	copyClip := cfg.CopyToClipboardEnabled()
	showQR := cfg.ShowQRCodeEnabled()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Copy URL to clipboard after upload?").
				Value(&copyClip),
			huh.NewConfirm().
				Title("Show QR code after upload?").
				Value(&showQR),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	cfg.CopyToClipboard = &copyClip
	cfg.ShowQRCode = &showQR

	if err := cfg.Write(); err != nil {
		return err
	}
	fmt.Println(tui.Success.Render("Preferences saved."))
	return nil
}

func pickProvider(cfg *config.Config, title string) (string, error) {
	return PickProvider(cfg, title)
}

// PickProvider shows an interactive menu to select a provider.
func PickProvider(cfg *config.Config, title string) (string, error) {
	var label string
	options := make([]huh.Option[string], len(cfg.Providers))
	for i, p := range cfg.Providers {
		options[i] = huh.NewOption(p.Label+" ("+p.Type+")", p.Label)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Value(&label),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return label, nil
}

// ReconfigureProvider re-authenticates a specific provider by label
func ReconfigureProvider(cfg *config.Config, label string) error {
	entry := cfg.FindByLabel(label)
	if entry == nil {
		return fmt.Errorf("provider %q not found", label)
	}

	fmt.Println(tui.Title.Render(fmt.Sprintf("Re-authenticating provider: %s (%s)", entry.Label, entry.Type)))
	fmt.Println("Your authentication has expired. Please authenticate again.")

	settings, err := runProviderForm(entry.Type, nil)
	if err != nil {
		return err
	}

	entry.Settings = settings
	if err := cfg.Write(); err != nil {
		return err
	}

	fmt.Println(tui.Success.Render(fmt.Sprintf("Provider %q re-authenticated successfully.", label)))
	return nil
}

// runProviderForm runs the appropriate form for a provider type and returns settings.
func runProviderForm(provType string, defaults map[string]string) (map[string]string, error) {
	switch provType {
	case "httpupload":
		form, fields := httpuploadForm(defaults)
		if err := form.Run(); err != nil {
			return nil, err
		}
		return fields.ToSettings(), nil

	case "nextcloud":
		form, fields := nextcloudForm(defaults)
		if err := form.Run(); err != nil {
			return nil, err
		}
		return fields.ToSettings(), nil

	case "seafile":
		form, fields := seafileForm(defaults)
		if err := form.Run(); err != nil {
			return nil, err
		}
		// Get token from seafile
		conf := seafile.Config{
			URL:              fields.URL,
			Username:         fields.Username,
			Password:         fields.Password,
			TwoFactorEnabled: fields.TwoFactorEnabled,
			OTP:              fields.OTP,
		}
		token, err := conf.GetToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get seafile token: %w", err)
		}
		conf.CreateLibrary(token)
		return map[string]string{
			"token":  token,
			"url":    fields.URL,
			"repoid": conf.RepoID,
		}, nil

	case "opendrive":
		form, fields := opendriveForm(defaults)
		if err := form.Run(); err != nil {
			return nil, err
		}
		return fields.ToSettings(), nil

	case "dropbox":
		conf := dropbox.OAuth2DropboxConfig()
		authURL := conf.AuthCodeURL("state", oauth2.SetAuthURLParam("token_access_type", "offline"))
		fmt.Printf("\n1. Go to %v\n", authURL)
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n\n")

		form, fields := dropboxForm()
		if err := form.Run(); err != nil {
			return nil, err
		}

		ctx := context.Background()
		token, err := conf.Exchange(ctx, fields.AuthCode)
		if err != nil {
			return nil, fmt.Errorf("dropbox token exchange failed: %w", err)
		}
		tokenB, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		return map[string]string{"token": string(tokenB)}, nil

	case "box":
		noteForm, proceed := oauthNoteForm("Box")
		if err := noteForm.Run(); err != nil {
			return nil, err
		}
		if !*proceed {
			return nil, fmt.Errorf("cancelled")
		}

		conf := box.OAuth2BoxConfig()
		result := RunOAuthFlowFixedPort(conf, "127.0.0.1:53682")
		if result.Err != nil {
			return nil, result.Err
		}
		tokenB, err := json.Marshal(result.Token)
		if err != nil {
			return nil, err
		}
		return map[string]string{"token": string(tokenB)}, nil

	case "googledrive":
		noteForm, proceed := oauthNoteForm("Google Drive")
		if err := noteForm.Run(); err != nil {
			return nil, err
		}
		if !*proceed {
			return nil, fmt.Errorf("cancelled")
		}

		conf := googledrive.OAuth2GoogleDriveConfig()
		result := RunOAuthFlow(conf, "")
		if result.Err != nil {
			return nil, result.Err
		}
		tokenB, err := json.Marshal(result.Token)
		if err != nil {
			return nil, err
		}
		return map[string]string{"googletoken": string(tokenB)}, nil

	default:
		return nil, fmt.Errorf("unknown provider type: %s", provType)
	}
}
