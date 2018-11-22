package main

import (
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

	"github.com/mitchellh/ioprogress"
)

const chunkSize int64 = 1 << 24

var (
	personalAppKey    = "40em1t168bc5aay"
	personalAppSecret = "mi55rqbz0rgb16f"
)

// Tokenmap example: { "token": "xxx" }
type TokenMap map[string]string

func oauth2DropboxConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     personalAppKey,
		ClientSecret: personalAppSecret,
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

// DropboxProvider implements a provider using dropbox sdk
type DropboxProvider struct {
	Config dropbox.Config
	token  string
}

func NewDropboxProvider(token string) *DropboxProvider {

	return &DropboxProvider{
		Config: dropbox.Config{
			Token:    token,
			LogLevel: dropbox.LogOff,
		},
	}
}

func (c *DropboxProvider) Upload(file *os.File, path string) (dst string, err error) {
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

func (c *DropboxProvider) GetLink(filepath string) (string, error) {
	share := sharing.New(c.Config)
	arg := sharing.NewCreateSharedLinkWithSettingsArg(filepath)

	res, err := share.CreateSharedLinkWithSettings(arg)
	if err != nil {
		if err.Error() == sharing.CreateSharedLinkWithSettingsErrorSharedLinkAlreadyExists {
			fmt.Println("exist bereits")
			return "", nil
		}
		//
		return "cannot create shared link", err
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
