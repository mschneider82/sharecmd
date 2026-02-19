package dropbox

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"golang.org/x/oauth2"
)

const chunkSize int64 = 1 << 24

var (
	ob = "ufsdii23n452u32iXXi8231aso0i1"
)

// TokenMap example: { "token": "xxx" }
type TokenMap map[string]string

// OAuth2DropboxConfig creates a oauth config
func OAuth2DropboxConfig() *oauth2.Config {
	hasher := sha1.New()
	hasher.Write([]byte(ob))
	ab := hasher.Sum(nil)[:16]
	o := obf{jkoq: []byte(ab)}
	endpoint := dropbox.OAuthEndpoint("")
	return &oauth2.Config{
		ClientID:     o.de("cJ21xYBoKXFzTY3vu1A3Hda4dp57jYMrTs1dbmdf9g=="),
		ClientSecret: o.de("Ziif+YX0+cnsKuO8P9ZBXhQwjs/IL/MwmdUnTbnZiQ=="),
		Endpoint:     endpoint,
	}
}

func readTokens(filePath string) (TokenMap, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var token TokenMap
	if json.Unmarshal(b, &token) != nil {
		return nil, err
	}
	return token, nil
}

func writeTokens(filePath string, tokens TokenMap) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Doesn't exist; lets create it
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return
		}
	}

	// At this point, file must exist. Lets (over)write it.
	b, err := json.Marshal(tokens)
	if err != nil {
		return
	}
	if err = os.WriteFile(filePath, b, 0600); err != nil {
		return
	}
}

// Provider implements a provider using dropbox sdk
type Provider struct {
	Config dropbox.Config
	token  string
}

// NewProvider creates a new Provider.
// tokenJSON can be either a plain access token string (legacy) or a
// JSON-encoded oauth2.Token (current format, supports automatic refresh).
func NewProvider(tokenJSON string) *Provider {
	cfg := dropbox.Config{LogLevel: dropbox.LogOff}

	var tok oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &tok); err == nil && tok.AccessToken != "" && tok.RefreshToken != "" {
		// Full token with refresh support — build an HTTP client that auto-refreshes.
		oauthCfg := OAuth2DropboxConfig()
		httpClient := oauthCfg.Client(context.Background(), &tok)
		cfg.Client = httpClient
	} else {
		// Legacy: plain access token string (will stop working once token expires).
		cfg.Token = tokenJSON
	}

	return &Provider{Config: cfg}
}

// Upload the file to dropbox
func (c *Provider) Upload(r io.Reader, filename string, size int64) (dst string, err error) {
	dst = "/" + filename

	delarg := files.NewDeleteArg(dst)
	dbx := files.New(c.Config)
	dbx.DeleteV2(delarg)

	uploadArg := files.NewUploadArg(dst)
	uploadArg.Mode.Tag = "overwrite"

	// The Dropbox API only accepts timestamps in UTC with second precision.
	t := time.Now().UTC().Round(time.Second)
	uploadArg.ClientModified = &t
	if size > chunkSize {
		return dst, uploadChunked(dbx, r, &uploadArg.CommitInfo, size)
	}

	if _, err = dbx.Upload(uploadArg, r); err != nil {
		return "", err
	}
	return dst, nil
}

// GetLink for file
func (c *Provider) GetLink(filepath string) (string, error) {
	share := sharing.New(c.Config)
	arg := sharing.NewCreateSharedLinkWithSettingsArg(filepath)

	res, err := share.CreateSharedLinkWithSettings(arg)
	if err != nil {
		return "", err
	}

	switch sl := res.(type) {
	case *sharing.FileLinkMetadata:
		return fixDropboxDownloadlink(sl.SharedLinkMetadata.Url), nil
	}

	return "", nil
}

// fixDropboxDownloadlink replaces dl=0 with dl=1 on for dropbox links to
// prevent signup popup and do direct downloading
func fixDropboxDownloadlink(link string) string {
	return link[:len(link)-1] + "1"
}

func uploadChunked(dbx files.Client, r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64) (err error) {
	res, err := dbx.UploadSessionStart(files.NewUploadSessionStartArg(),
		&io.LimitedReader{R: r, N: chunkSize})
	if err != nil {
		return
	}

	written := chunkSize

	for (sizeTotal - written) > chunkSize {
		cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
		args := files.NewUploadSessionAppendArg(cursor)

		err = dbx.UploadSessionAppendV2(args, &io.LimitedReader{R: r, N: chunkSize})
		if err != nil {
			return
		}
		written += chunkSize
	}

	cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
	args := files.NewUploadSessionFinishArg(cursor, commitInfo)

	if _, err = dbx.UploadSessionFinish(args, r); err != nil {
		return
	}

	return
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
