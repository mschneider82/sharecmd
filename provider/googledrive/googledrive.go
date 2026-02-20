package googledrive

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// https://developers.google.com/drive/api/v3/quickstart/go

// Provider implements a provider
type Provider struct {
	Config         *oauth2.Config
	token          *oauth2.Token
	tokenSource    oauth2.TokenSource
	onTokenRefresh func(newToken *oauth2.Token)
}

var mimeExtentions = map[string]string{
	".epub":  "application/epub+zip",
	".json":  "application/json",
	".doc":   "application/msword",
	".pdf":   "application/pdf",
	".rtf":   "application/rtf",
	".xls":   "application/vnd.ms-excel",
	".odp":   "application/vnd.oasis.opendocument.presentation",
	".ods":   "application/vnd.oasis.opendocument.spreadsheet",
	".odt":   "application/vnd.oasis.opendocument.text",
	".pptx":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".xlsx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".wmf":   "application/x-msmetafile",
	".zip":   "application/zip",
	".bmp":   "image/bmp",
	".jpg":   "image/jpeg",
	".pjpeg": "image/pjpeg",
	".png":   "image/png",
	".svg":   "image/svg+xml",
	".csv":   "text/csv",
	".html":  "text/html",
	".txt":   "text/plain",
	".tsv":   "text/tab-separated-values",
}

var (
	ob = "ufsdii23n452u32iXXi8231aso0i1"
	y  = "MmciVUipVqmm4Chej+dVMxwUumsQDTq3G6Qkv7lhR366CaVac3eD1w=="
)

// OAuth2GoogleDriveConfig ...
func OAuth2GoogleDriveConfig() *oauth2.Config {
	hasher := sha1.New()
	hasher.Write([]byte(ob))
	ab := hasher.Sum(nil)[:16]
	k := obf{jkoq: []byte(ab)}
	x := k.de(y)
	b := []byte(`{"installed":{"client_id":"26115953275-7971erj532s8d98vlso25467iudikbvf.apps.googleusercontent.com","project_id":"sharecmd-223413","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://www.googleapis.com/oauth2/v3/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"` + x + `","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`)
	// config, err := google.ConfigFromJSON(b, drive.DriveMetadataScope)
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}

// NewProvider creates a new Provider
func NewProvider(token string) *Provider {
	tok := &oauth2.Token{}
	err := json.Unmarshal([]byte(token), tok)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err)
	}

	cfg := OAuth2GoogleDriveConfig()
	p := &Provider{
		token:  tok,
		Config: cfg,
	}
	p.tokenSource = &notifyingTokenSource{
		src: cfg.TokenSource(context.Background(), tok),
		onRefresh: func(newToken *oauth2.Token) {
			p.token = newToken
			if p.onTokenRefresh != nil {
				p.onTokenRefresh(newToken)
			}
		},
	}
	return p
}

// SetTokenRefreshCallback sets a callback that's invoked when the token is refreshed
func (c *Provider) SetTokenRefreshCallback(callback func(newToken *oauth2.Token)) {
	c.onTokenRefresh = callback
}

// GetCurrentToken returns the current (possibly refreshed) token
func (c *Provider) GetCurrentToken() *oauth2.Token {
	return c.token
}

func (c *Provider) getClient() *http.Client {
	return oauth2.NewClient(context.Background(), c.tokenSource)
}

// notifyingTokenSource wraps a TokenSource and calls a callback on token refresh
type notifyingTokenSource struct {
	src       oauth2.TokenSource
	onRefresh func(*oauth2.Token)
}

func (n *notifyingTokenSource) Token() (*oauth2.Token, error) {
	token, err := n.src.Token()
	if err != nil {
		return nil, err
	}
	if n.onRefresh != nil {
		n.onRefresh(token)
	}
	return token, nil
}

// Upload the file
func (c *Provider) Upload(r io.Reader, filename string, size int64) (fileID string, err error) {
	client := c.getClient()
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	parendID := getOrCreateFolder(srv, "sharecmd")

	fileext := filepath.Ext(filename)

	f := &drive.File{
		Name:    filename,
		Parents: []string{parendID},
	}
	if mimeExtentions[fileext] != "" {
		f.MimeType = mimeExtentions[fileext]
	}
	result, err := srv.Files.Create(f).Media(r).Do()
	if err != nil {
		return "", err
	}
	return result.Id, nil
}

// GetLink for fileid
func (c *Provider) GetLink(filepath string) (string, error) {
	fileID := filepath

	client := c.getClient()
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}

	_, err = srv.Permissions.Create(fileID, permission).Do()
	if err != nil {
		return "", err
	}

	link := fmt.Sprintf("https://drive.google.com/file/d/%s/view?usp=sharing", fileID)

	return link, nil
}

func getOrCreateFolder(d *drive.Service, folderName string) string {
	folderID := ""
	if folderName == "" {
		folderName = "sharecmd"
	}
	q := fmt.Sprintf("name=\"%s\" and mimeType=\"application/vnd.google-apps.folder\"", folderName)

	r, err := d.Files.List().Q(q).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve foldername: %s", err.Error())
	}
	if len(r.Files) > 0 {
		folderID = r.Files[0].Id
	} else {
		// no folder found create new
		log.Printf("Folder not found. Create new folder : %s\n", folderName)
		f := &drive.File{Name: folderName, Description: "Auto Create by sharecmd", MimeType: "application/vnd.google-apps.folder"}
		r, err := d.Files.Create(f).Do()
		if err != nil {
			log.Printf("An error occurred when create folder: %v\n", err)
		}
		folderID = r.Id
	}
	return folderID
}

type obf struct {
	jkoq []byte
}

func (b *obf) en(plain string) string {
	text := []byte(plain)
	block, err := aes.NewCipher(b.jkoq)
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func (b *obf) de(b64 string) string {
	text, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(err)
	}
	block, err := aes.NewCipher(b.jkoq)
	if err != nil {
		panic(err)
	}
	if len(text) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	return string(text)
}
