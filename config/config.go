package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
)

// ProviderEntry holds a single provider configuration.
type ProviderEntry struct {
	Label    string            `json:"label"`
	Type     string            `json:"type"`
	Settings map[string]string `json:"settings"`
}

// Config is the v2 configuration format supporting multiple providers.
type Config struct {
	Version         int             `json:"version"`
	Active          string          `json:"active"`
	Providers       []ProviderEntry `json:"providers"`
	CopyToClipboard *bool           `json:"copy_to_clipboard,omitempty"`
	ShowQRCode      *bool           `json:"show_qr_code,omitempty"`
	Path            string          `json:"-"`
}

// CopyToClipboardEnabled returns whether clipboard copy is enabled (default: true).
func (c *Config) CopyToClipboardEnabled() bool {
	if c.CopyToClipboard == nil {
		return true
	}
	return *c.CopyToClipboard
}

// ShowQRCodeEnabled returns whether QR code display is enabled (default: true).
func (c *Config) ShowQRCodeEnabled() bool {
	if c.ShowQRCode == nil {
		return true
	}
	return *c.ShowQRCode
}

// configV1 is the legacy single-provider format (version 1 / no version field).
type configV1 struct {
	Provider             string            `json:"provider"`
	ProviderSettings     map[string]string `json:"providersettings"`
	URLShortenerProvider string            `json:"urlshortenerprovider,omitempty"`
	URLShortenerSettings map[string]string `json:"urlshortenersettings,omitempty"`
}

// UserHomeDir returns the user's home directory.
func UserHomeDir() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	return os.Getenv(env)
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	return UserHomeDir() + "/.config/sharecmd/config.json"
}

// ActiveProvider returns the currently active ProviderEntry, or nil.
func (c *Config) ActiveProvider() *ProviderEntry {
	return c.FindByLabel(c.Active)
}

// FindByLabel returns the provider with the given label, or nil.
func (c *Config) FindByLabel(label string) *ProviderEntry {
	for i := range c.Providers {
		if c.Providers[i].Label == label {
			return &c.Providers[i]
		}
	}
	return nil
}

// AddProvider appends a new provider entry.
func (c *Config) AddProvider(entry ProviderEntry) {
	c.Providers = append(c.Providers, entry)
}

// RemoveProvider removes the provider with the given label.
// If the removed provider was active, Active is cleared.
func (c *Config) RemoveProvider(label string) {
	for i := range c.Providers {
		if c.Providers[i].Label == label {
			c.Providers = append(c.Providers[:i], c.Providers[i+1:]...)
			if c.Active == label {
				c.Active = ""
			}
			return
		}
	}
}

// SetActive sets the active provider by label. Returns error if not found.
func (c *Config) SetActive(label string) error {
	if c.FindByLabel(label) == nil {
		return fmt.Errorf("provider %q not found", label)
	}
	c.Active = label
	return nil
}

// ProviderLabels returns a list of all configured provider labels.
func (c *Config) ProviderLabels() []string {
	labels := make([]string, len(c.Providers))
	for i, p := range c.Providers {
		labels[i] = p.Label
	}
	return labels
}

// Write saves the config to disk at its Path.
func (c *Config) Write() error {
	err := os.MkdirAll(path.Dir(c.Path), 0o700)
	if err != nil {
		return err
	}
	output, err := os.Create(c.Path)
	if err != nil {
		return err
	}
	defer output.Close()

	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

// LoadConfig loads the config from disk, auto-migrating v1 format.
func LoadConfig(filepath string) (*Config, error) {
	content, err := os.ReadFile(filepath)
	if os.IsNotExist(err) {
		return &Config{
			Version:   2,
			Path:      filepath,
			Providers: []ProviderEntry{},
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// Peek at the version field.
	var probe struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(content, &probe); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}

	if probe.Version >= 2 {
		cfg := &Config{}
		if err := json.Unmarshal(content, cfg); err != nil {
			return nil, err
		}
		cfg.Path = filepath
		return cfg, nil
	}

	// v1 migration
	return migrateV1(content, filepath)
}

func migrateV1(data []byte, filepath string) (*Config, error) {
	var v1 configV1
	if err := json.Unmarshal(data, &v1); err != nil {
		return nil, fmt.Errorf("failed to parse v1 config: %w", err)
	}

	cfg := &Config{
		Version:   2,
		Path:      filepath,
		Providers: []ProviderEntry{},
	}

	if v1.Provider != "" {
		settings := v1.ProviderSettings
		if settings == nil {
			settings = make(map[string]string)
		}
		entry := ProviderEntry{
			Label:    v1.Provider,
			Type:     v1.Provider,
			Settings: settings,
		}
		cfg.Providers = append(cfg.Providers, entry)
		cfg.Active = v1.Provider
	}

	return cfg, nil
}

// LookupConfig loads config from the given path (or default).
func LookupConfig(configfilepath string) (*Config, error) {
	if configfilepath == "" {
		configfilepath = DefaultConfigPath()
	}
	return LoadConfig(configfilepath)
}
