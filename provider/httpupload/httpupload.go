package httpupload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"
)

// Provider uploads files via HTTP PUT to a base URL.
// The final URL is baseURL + filename.
// Header values support Go template functions for dynamic values:
//
//	{{now "2006-01-02"}}          → today formatted
//	{{addDays 7 "2006-01-02"}}   → today + N days formatted
type Provider struct {
	BaseURL string
	Headers map[string]string
}

var tmplFuncs = template.FuncMap{
	"now": func(layout string) string {
		return time.Now().Format(layout)
	},
	"addDays": func(days int, layout string) string {
		return time.Now().AddDate(0, 0, days).Format(layout)
	},
}

// NewProvider creates a new HTTP upload provider.
// headersJSON is a JSON-encoded map[string]string (may be empty or "{}").
func NewProvider(baseURL, headersJSON string) *Provider {
	headers := make(map[string]string)
	if headersJSON != "" {
		json.Unmarshal([]byte(headersJSON), &headers) //nolint:errcheck
	}
	return &Provider{
		BaseURL: strings.TrimRight(baseURL, "/") + "/",
		Headers: headers,
	}
}

// renderValue evaluates Go template expressions in a header value.
// If the value contains no templates, it is returned as-is.
func renderValue(val string) string {
	if !strings.Contains(val, "{{") {
		return val
	}
	t, err := template.New("").Funcs(tmplFuncs).Parse(val)
	if err != nil {
		return val
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, nil); err != nil {
		return val
	}
	return buf.String()
}

// Upload PUTs the file content to baseURL/filename.
func (p *Provider) Upload(r io.Reader, filename string, size int64) (string, error) {
	url := p.BaseURL + filename

	req, err := http.NewRequest("PUT", url, r)
	if err != nil {
		return "", err
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", "application/octet-stream")

	for k, v := range p.Headers {
		req.Header.Set(k, renderValue(v))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP PUT failed (%d): %s", resp.StatusCode, string(body))
	}

	return url, nil
}

// GetLink returns the URL that was already constructed during Upload.
func (p *Provider) GetLink(fileURL string) (string, error) {
	return fileURL, nil
}
