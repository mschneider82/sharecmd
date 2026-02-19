package setup

import (
	"github.com/charmbracelet/huh"
)

// ProviderTypes lists all available provider types.
var ProviderTypes = []string{"httpupload", "nextcloud", "dropbox", "googledrive", "box", "opendrive", "seafile"}

// NextcloudFields holds the form field values for a nextcloud provider.
type NextcloudFields struct {
	URL                   string
	Username              string
	Password              string
	LinkShareWithPassword bool
	RandomPasswordChars   string
}

func (f *NextcloudFields) ToSettings() map[string]string {
	lswp := "false"
	if f.LinkShareWithPassword {
		lswp = "true"
	}
	return map[string]string{
		"url":                   f.URL,
		"username":              f.Username,
		"password":              f.Password,
		"linkShareWithPassword": lswp,
		"randomPasswordChars":   f.RandomPasswordChars,
	}
}

func nextcloudForm(defaults map[string]string) (*huh.Form, *NextcloudFields) {
	f := &NextcloudFields{
		URL:                   getDefault(defaults, "url", ""),
		Username:              getDefault(defaults, "username", ""),
		Password:              getDefault(defaults, "password", ""),
		LinkShareWithPassword: getDefault(defaults, "linkShareWithPassword", "") == "true",
		RandomPasswordChars:   getDefault(defaults, "randomPasswordChars", "32"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Nextcloud URL").
				Description("e.g. https://example.com").
				Value(&f.URL),
			huh.NewInput().
				Title("Username").
				Value(&f.Username),
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&f.Password),
			huh.NewConfirm().
				Title("Password-protected link shares?").
				Value(&f.LinkShareWithPassword),
			huh.NewInput().
				Title("Random password length").
				Value(&f.RandomPasswordChars),
		),
	)

	return form, f
}

// SeafileFields holds the form field values for a seafile provider.
type SeafileFields struct {
	URL              string
	Username         string
	Password         string
	TwoFactorEnabled bool
	OTP              string
}

func seafileForm(defaults map[string]string) (*huh.Form, *SeafileFields) {
	f := &SeafileFields{
		URL:      getDefault(defaults, "url", ""),
		Username: getDefault(defaults, "username", ""),
		Password: getDefault(defaults, "password", ""),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Seafile URL").
				Description("e.g. https://seacloud.cc").
				Value(&f.URL),
			huh.NewInput().
				Title("Username").
				Value(&f.Username),
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&f.Password),
			huh.NewConfirm().
				Title("Two-factor auth enabled?").
				Value(&f.TwoFactorEnabled),
			huh.NewInput().
				Title("OTP Token").
				Description("Only needed if 2FA is enabled").
				Value(&f.OTP),
		),
	)

	return form, f
}

// OpendriveFields holds the form field values for an opendrive provider.
type OpendriveFields struct {
	User     string
	Password string
}

func (f *OpendriveFields) ToSettings() map[string]string {
	return map[string]string{
		"user": f.User,
		"pass": f.Password,
	}
}

func opendriveForm(defaults map[string]string) (*huh.Form, *OpendriveFields) {
	f := &OpendriveFields{
		User:     getDefault(defaults, "user", ""),
		Password: getDefault(defaults, "pass", ""),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Value(&f.User),
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&f.Password),
		),
	)
	return form, f
}

// DropboxFields holds the authorization code for dropbox.
type DropboxFields struct {
	AuthCode string
}

func dropboxForm() (*huh.Form, *DropboxFields) {
	f := &DropboxFields{}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Dropbox Authorization").
				Description("A browser window will open.\nClick \"Allow\" and paste the authorization code below."),
			huh.NewInput().
				Title("Authorization Code").
				Value(&f.AuthCode),
		),
	)
	return form, f
}

func oauthNoteForm(providerName string) (*huh.Form, *bool) {
	proceed := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(providerName + " Authorization").
				Description("A browser window will open for authorization.\nPress Enter to continue..."),
			huh.NewConfirm().
				Title("Open browser?").
				Affirmative("Continue").
				Negative("Cancel").
				Value(&proceed),
		),
	)
	return form, &proceed
}

func labelForm(defaultLabel string) (*huh.Form, *string) {
	label := defaultLabel
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Label").
				Description("A name for this provider configuration").
				Value(&label),
		),
	)
	return form, &label
}

// HTTPUploadFields holds the form field values for an httpupload provider.
type HTTPUploadFields struct {
	URL     string
	Headers string
}

func (f *HTTPUploadFields) ToSettings() map[string]string {
	return map[string]string{
		"url":     f.URL,
		"headers": f.Headers,
	}
}

func httpuploadForm(defaults map[string]string) (*huh.Form, *HTTPUploadFields) {
	f := &HTTPUploadFields{
		URL:     getDefault(defaults, "url", ""),
		Headers: getDefault(defaults, "headers", "{}"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Base URL").
				Description("Files are PUT to <url>/<filename>\ne.g. https://example.com/uploads/").
				Value(&f.URL),
			huh.NewText().
				Title("Custom HTTP Headers (JSON)").
				Description("e.g. {\"Authorization\": \"Bearer token\"}\nTemplate functions: {{now \"2006-01-02\"}}, {{addDays 7 \"2006-01-02\"}}").
				Value(&f.Headers),
		),
	)
	return form, f
}

func getDefault(m map[string]string, key, fallback string) string {
	if m == nil {
		return fallback
	}
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return fallback
}
