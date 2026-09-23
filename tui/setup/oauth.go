package setup

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/oauth2"
)

// OAuthResult holds the result of an OAuth flow.
type OAuthResult struct {
	Token *oauth2.Token
	Err   error
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}

// RunOAuthFlowWithManualFallback runs OAuth with both automatic callback and manual URL entry support.
// This allows setup to work both locally (browser opens automatically) and remotely (SSH, no browser).
// If listenAddr is empty, a random port is used. For fixed ports (e.g. Box), pass the specific address.
func RunOAuthFlowWithManualFallback(oauthConf *oauth2.Config, listenAddr string, fixedRedirect bool) OAuthResult {
	if listenAddr == "" {
		listenAddr = "127.0.0.1:0"
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return OAuthResult{Err: fmt.Errorf("failed to start local OAuth server: %w", err)}
	}
	port := ln.Addr().(*net.TCPAddr).Port

	// Update redirect URL to match the actual port (unless using a fixed redirect URL).
	if !fixedRedirect {
		oauthConf.RedirectURL = fmt.Sprintf("http://localhost:%d", port)
	}

	authURL := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Channel for receiving the auth code from browser callback
	codeCh := make(chan string, 1)

	// Start HTTP server for automatic callback
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "<html><body><h2>Authorization successful!</h2><p>You can close this tab.</p></body></html>")
			select {
			case codeCh <- code:
			default:
			}
		}
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	// Try to open browser (non-blocking, may fail on remote systems)
	openBrowser(authURL)

	// Create cancelable context for the form
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Show huh form with auth URL and manual input option
	var manualInput string
	// Escape underscores in URL to prevent markdown interpretation
	escapedAuthURL := strings.ReplaceAll(authURL, "_", "\\_")
	escapedRedirectURL := strings.ReplaceAll(oauthConf.RedirectURL, "_", "\\_")
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("OAuth Authorization").
				Description(fmt.Sprintf(
					"1. Open this URL in your browser:\n\n   %s\n\n"+
						"2. Authorize the application\n\n"+
						"3. If browser opened automatically, just wait.\n"+
						"   Otherwise, paste the redirect URL below.",
					escapedAuthURL,
				)),
			huh.NewInput().
				Title("Redirect URL (optional)").
				Description(fmt.Sprintf("Leave empty to wait for browser. URL looks like: %s?code=...", escapedRedirectURL)).
				Value(&manualInput),
		),
	).WithAccessible(false)

	// Run form in goroutine so we can also listen for browser callback
	formDone := make(chan error, 1)
	go func() {
		formDone <- form.RunWithContext(ctx)
	}()

	// Wait for either: form completion or browser callback
	var authcode string
	select {
	case code := <-codeCh:
		// Browser callback received - cancel the form
		cancel()
		// Wait for form goroutine to finish
		<-formDone
		authcode = code
		fmt.Println("\n✓ Authorization received from browser!")
	case err := <-formDone:
		if err != nil && err != context.Canceled {
			return OAuthResult{Err: err}
		}
		// Form completed - check if user entered something
		if manualInput != "" {
			authcode = extractCodeFromInput(manualInput)
			if authcode == "" {
				return OAuthResult{Err: fmt.Errorf("could not extract authorization code from input")}
			}
		} else {
			// User pressed Enter without input - wait for browser callback
			fmt.Println("Waiting for browser authorization...")
			authcode = <-codeCh
			fmt.Println("\n✓ Authorization received from browser!")
		}
	}

	tok, err := oauthConf.Exchange(context.TODO(), authcode)
	if err != nil {
		return OAuthResult{Err: fmt.Errorf("token exchange failed: %w", err)}
	}
	return OAuthResult{Token: tok}
}

// extractCodeFromInput extracts the authorization code from user input.
// Input can be a full redirect URL or just the code itself.
func extractCodeFromInput(input string) string {
	input = strings.TrimSpace(input)

	// Try parsing as URL first
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		if u, err := url.Parse(input); err == nil {
			if code := u.Query().Get("code"); code != "" {
				return code
			}
		}
	}

	// If not a URL with code parameter, treat the whole input as the code
	// (in case user copied just the code value)
	if len(input) > 10 && !strings.Contains(input, " ") {
		return input
	}

	return ""
}
