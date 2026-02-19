package box

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	authURL     = "https://app.box.com/api/oauth2/authorize"
	tokenURL    = "https://api.box.com/oauth2/token"
	apiBase     = "https://api.box.com/2.0"
	uploadBase  = "https://upload.box.com/api/2.0"
	redirectURL = "http://localhost:53682"
)

var (
	ob  = "bx7p2n9q4m8k3r1"
	eID = "BWz7JR6WsAiyig3hy6/ywju5vCmYUEp2F4cJztT7QpXfelhUvbLUJXgnzizkjnsp"
	eSc = "ddwZ5dGf8Lsny71gP3jzezvKB/4mrS/GneJVtcNQa60Ak2W0N5i6gs0h1dN1iLoc"
)

// Provider implements a Box provider
type Provider struct {
	config *oauth2.Config
	token  *oauth2.Token
}

// OAuth2BoxConfig returns the OAuth2 config for Box
func OAuth2BoxConfig() *oauth2.Config {
	k := newObf(ob)
	return &oauth2.Config{
		ClientID:     k.de(eID),
		ClientSecret: k.de(eSc),
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
}

// NewProvider creates a new Box Provider from a JSON-encoded oauth2.Token
func NewProvider(token string) *Provider {
	tok := &oauth2.Token{}
	if err := json.Unmarshal([]byte(token), tok); err != nil {
		log.Fatalf("Unable to parse Box token: %v", err)
	}
	return &Provider{token: tok, config: OAuth2BoxConfig()}
}

func (p *Provider) httpClient() *http.Client {
	return p.config.Client(context.Background(), p.token)
}

// Upload uploads a file to Box inside a "sharecmd" folder and returns the file ID
func (p *Provider) Upload(r io.Reader, filename string, size int64) (string, error) {
	client := p.httpClient()

	// Buffer content so it can be reused if we need to upload a new version
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	folderID, err := getOrCreateFolder(client, "sharecmd")
	if err != nil {
		return "", fmt.Errorf("folder: %w", err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	attrs := fmt.Sprintf(`{"name":%q,"parent":{"id":%q}}`, filename, folderID)
	if err := mw.WriteField("attributes", attrs); err != nil {
		return "", err
	}
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(content); err != nil {
		return "", err
	}
	mw.Close()

	req, err := http.NewRequest("POST", uploadBase+"/files/content", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		// File already exists — parse the conflicting file ID and upload a new version
		var conflict struct {
			ContextInfo struct {
				Conflicts struct {
					ID string `json:"id"`
				} `json:"conflicts"`
			} `json:"context_info"`
		}
		b, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(b, &conflict); err != nil || conflict.ContextInfo.Conflicts.ID == "" {
			return "", fmt.Errorf("upload conflict but could not parse existing file ID: %s", string(b))
		}
		return p.uploadNewVersion(client, conflict.ContextInfo.Conflicts.ID, filename, content)
	}

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Entries []struct {
			ID string `json:"id"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Entries) == 0 {
		return "", fmt.Errorf("upload response contained no file entries")
	}
	return result.Entries[0].ID, nil
}

// uploadNewVersion replaces an existing file on Box with new content
func (p *Provider) uploadNewVersion(client *http.Client, fileID, filename string, content []byte) (string, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	attrs := fmt.Sprintf(`{"name":%q}`, filename)
	if err := mw.WriteField("attributes", attrs); err != nil {
		return "", err
	}
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(content); err != nil {
		return "", err
	}
	mw.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/files/%s/content", uploadBase, fileID), &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload new version failed (%d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Entries []struct {
			ID string `json:"id"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Entries) == 0 {
		return "", fmt.Errorf("upload new version response contained no file entries")
	}
	return result.Entries[0].ID, nil
}

// GetLink creates a public shared link for the given file ID and returns the URL
func (p *Provider) GetLink(fileID string) (string, error) {
	client := p.httpClient()

	payload := `{"shared_link":{"access":"open"}}`
	req, err := http.NewRequest("PUT",
		fmt.Sprintf("%s/files/%s?fields=shared_link", apiBase, fileID),
		bytes.NewBufferString(payload),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("shared link failed (%d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		SharedLink struct {
			URL string `json:"url"`
		} `json:"shared_link"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.SharedLink.URL, nil
}

func getOrCreateFolder(client *http.Client, name string) (string, error) {
	// Search in root folder (id "0")
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/folders/0/items?fields=id,name,type&limit=1000", apiBase),
		nil,
	)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var listing struct {
		Entries []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return "", err
	}
	for _, e := range listing.Entries {
		if e.Type == "folder" && e.Name == name {
			return e.ID, nil
		}
	}

	// Create folder
	payload := fmt.Sprintf(`{"name":%q,"parent":{"id":"0"}}`, name)
	req, err = http.NewRequest("POST", apiBase+"/folders", bytes.NewBufferString(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var folder struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return "", err
	}
	return folder.ID, nil
}

type obf struct {
	key []byte
}

func newObf(seed string) obf {
	h := sha1.New()
	h.Write([]byte(seed))
	return obf{key: h.Sum(nil)[:16]}
}

func (b obf) de(enc string) string {
	text, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		panic(err)
	}
	block, err := aes.NewCipher(b.key)
	if err != nil {
		panic(err)
	}
	if len(text) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cipher.NewCFBDecrypter(block, iv).XORKeyStream(text, text)
	return string(text)
}
