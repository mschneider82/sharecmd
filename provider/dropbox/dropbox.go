package dropbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"github.com/dustin/go-humanize"
	"golang.org/x/oauth2"

	//"github.com/mitchellh/ioprogress"
	"github.com/coreos/ioprogress"
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
	return &oauth2.Config{
		ClientID:     o.de("cJ21xYBoKXFzTY3vu1A3Hda4dp57jYMrTs1dbmdf9g=="),
		ClientSecret: o.de("Ziif+YX0+cnsKuO8P9ZBXhQwjs/IL/MwmdUnTbnZiQ=="),
		Endpoint:     dropbox.OAuthEndpoint(".dropboxapi.com"),
	}
}

func readTokens(filePath string) (TokenMap, error) {
	b, err := ioutil.ReadFile(filePath)
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
	if err = ioutil.WriteFile(filePath, b, 0600); err != nil {
		return
	}
}

// Provider implements a provider using dropbox sdk
type Provider struct {
	Config dropbox.Config
	token  string
}

// NewProvider creates a new Provider
func NewProvider(token string) *Provider {
	return &Provider{
		Config: dropbox.Config{
			Token:    token,
			LogLevel: dropbox.LogOff,
		},
	}
}

// Upload the file to dropbox
func (c *Provider) Upload(file *os.File, path string) (dst string, err error) {
	//
	if path == "" {
		path = "/"
	}
	dst = path + filepath.Base(file.Name())
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	delarg := files.NewDeleteArg(dst)
	dbx := files.New(c.Config)
	dbx.DeleteV2(delarg)

	progressbar := &ioprogress.Reader{
		Reader: file,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: fileInfo.Size(),
	}

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"

	// The Dropbox API only accepts timestamps in UTC with second precision.
	commitInfo.ClientModified = time.Now().UTC().Round(time.Second)
	//	dbx := files.New(c.Config)
	if fileInfo.Size() > chunkSize {
		return "", uploadChunked(dbx, progressbar, commitInfo, fileInfo.Size())
	}

	if _, err = dbx.Upload(commitInfo, progressbar); err != nil {
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
