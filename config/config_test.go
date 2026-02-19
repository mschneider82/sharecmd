package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateV1(t *testing.T) {
	v1JSON := `{
		"provider": "nextcloud",
		"providersettings": {
			"url": "https://example.com",
			"username": "user",
			"password": "pass",
			"linkShareWithPassword": "true",
			"randomPasswordChars": "32"
		}
	}`

	cfg, err := migrateV1([]byte(v1JSON), "/tmp/test.json")
	if err != nil {
		t.Fatalf("migrateV1: %v", err)
	}

	if cfg.Version != 2 {
		t.Errorf("expected version 2, got %d", cfg.Version)
	}
	if cfg.Active != "nextcloud" {
		t.Errorf("expected active=nextcloud, got %q", cfg.Active)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(cfg.Providers))
	}

	p := cfg.Providers[0]
	if p.Label != "nextcloud" {
		t.Errorf("expected label=nextcloud, got %q", p.Label)
	}
	if p.Type != "nextcloud" {
		t.Errorf("expected type=nextcloud, got %q", p.Type)
	}
	if p.Settings["url"] != "https://example.com" {
		t.Errorf("expected url=https://example.com, got %q", p.Settings["url"])
	}
}

func TestMigrateV1Empty(t *testing.T) {
	v1JSON := `{}`

	cfg, err := migrateV1([]byte(v1JSON), "/tmp/test.json")
	if err != nil {
		t.Fatalf("migrateV1: %v", err)
	}

	if cfg.Version != 2 {
		t.Errorf("expected version 2, got %d", cfg.Version)
	}
	if cfg.Active != "" {
		t.Errorf("expected empty active, got %q", cfg.Active)
	}
	if len(cfg.Providers) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(cfg.Providers))
	}
}

func TestLoadConfigV2(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	cfg := &Config{
		Version: 2,
		Active:  "my-dropbox",
		Providers: []ProviderEntry{
			{Label: "my-dropbox", Type: "dropbox", Settings: map[string]string{"token": "abc"}},
		},
		Path: p,
	}
	if err := cfg.Write(); err != nil {
		t.Fatalf("Write: %v", err)
	}

	loaded, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Version != 2 {
		t.Errorf("expected version 2, got %d", loaded.Version)
	}
	if loaded.Active != "my-dropbox" {
		t.Errorf("expected active=my-dropbox, got %q", loaded.Active)
	}
	if len(loaded.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(loaded.Providers))
	}
}

func TestLoadConfigAutoMigrate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	v1 := map[string]interface{}{
		"provider":         "dropbox",
		"providersettings": map[string]string{"token": "xyz"},
	}
	data, _ := json.Marshal(v1)
	os.WriteFile(p, data, 0o600)

	loaded, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Version != 2 {
		t.Errorf("expected version 2, got %d", loaded.Version)
	}
	if loaded.Active != "dropbox" {
		t.Errorf("expected active=dropbox, got %q", loaded.Active)
	}
	if loaded.Providers[0].Settings["token"] != "xyz" {
		t.Errorf("expected token=xyz, got %q", loaded.Providers[0].Settings["token"])
	}
}

func TestLoadConfigNotExist(t *testing.T) {
	cfg, err := LoadConfig("/tmp/nonexistent-sharecmd-test.json")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Version != 2 {
		t.Errorf("expected version 2, got %d", cfg.Version)
	}
	if len(cfg.Providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(cfg.Providers))
	}
}

func TestConfigHelpers(t *testing.T) {
	cfg := &Config{
		Version:   2,
		Providers: []ProviderEntry{},
		Path:      "/tmp/test.json",
	}

	cfg.AddProvider(ProviderEntry{Label: "work-nc", Type: "nextcloud", Settings: map[string]string{"url": "https://nc.work"}})
	cfg.AddProvider(ProviderEntry{Label: "personal-db", Type: "dropbox", Settings: map[string]string{"token": "abc"}})

	if len(cfg.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(cfg.Providers))
	}

	p := cfg.FindByLabel("work-nc")
	if p == nil {
		t.Fatal("FindByLabel(work-nc) returned nil")
	}
	if p.Type != "nextcloud" {
		t.Errorf("expected type=nextcloud, got %q", p.Type)
	}

	if err := cfg.SetActive("work-nc"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}
	if cfg.Active != "work-nc" {
		t.Errorf("expected active=work-nc, got %q", cfg.Active)
	}

	ap := cfg.ActiveProvider()
	if ap == nil || ap.Label != "work-nc" {
		t.Error("ActiveProvider() mismatch")
	}

	if err := cfg.SetActive("nonexistent"); err == nil {
		t.Error("SetActive should fail for nonexistent label")
	}

	cfg.RemoveProvider("work-nc")
	if len(cfg.Providers) != 1 {
		t.Fatalf("expected 1 provider after remove, got %d", len(cfg.Providers))
	}
	if cfg.Active != "" {
		t.Errorf("expected active cleared after removing active provider, got %q", cfg.Active)
	}

	labels := cfg.ProviderLabels()
	if len(labels) != 1 || labels[0] != "personal-db" {
		t.Errorf("unexpected labels: %v", labels)
	}
}
