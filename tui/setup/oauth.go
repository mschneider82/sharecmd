package setup

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"golang.org/x/oauth2"
)

// OAuthResult holds the result of an OAuth flow.
type OAuthResult struct {
	Token *oauth2.Token
	Err   error
}

// RunOAuthFlow starts an HTTP loopback server, opens the browser to authURL,
// waits for the callback, exchanges the code, and returns the token.
// If listenAddr is empty, a random port is used.
func RunOAuthFlow(oauthConf *oauth2.Config, listenAddr string) OAuthResult {
	if listenAddr == "" {
		listenAddr = "127.0.0.1:0"
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return OAuthResult{Err: fmt.Errorf("failed to start local OAuth server: %w", err)}
	}
	port := ln.Addr().(*net.TCPAddr).Port

	// Update redirect URL to match the actual port.
	oauthConf.RedirectURL = fmt.Sprintf("http://localhost:%d", port)

	authURL := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	codeCh := make(chan string, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "<html><body><h2>Authorization successful!</h2><p>You can close this tab.</p></body></html>")
			codeCh <- code
		}
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln) //nolint:errcheck

	openBrowser(authURL)

	authcode := <-codeCh
	srv.Close()

	tok, err := oauthConf.Exchange(context.TODO(), authcode)
	if err != nil {
		return OAuthResult{Err: fmt.Errorf("token exchange failed: %w", err)}
	}
	return OAuthResult{Token: tok}
}

// RunOAuthFlowFixedPort runs OAuth on a specific address (e.g. "127.0.0.1:53682" for Box).
func RunOAuthFlowFixedPort(oauthConf *oauth2.Config, addr string) OAuthResult {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return OAuthResult{Err: fmt.Errorf("failed to start local OAuth server on %s: %w", addr, err)}
	}

	authURL := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	codeCh := make(chan string, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "<html><body><h2>Authorization successful!</h2><p>You can close this tab.</p></body></html>")
			codeCh <- code
		}
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln) //nolint:errcheck

	openBrowser(authURL)

	authcode := <-codeCh
	srv.Close()

	tok, err := oauthConf.Exchange(context.TODO(), authcode)
	if err != nil {
		return OAuthResult{Err: fmt.Errorf("token exchange failed: %w", err)}
	}
	return OAuthResult{Token: tok}
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
