package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	humanize "github.com/dustin/go-humanize"
	//"github.com/mitchellh/ioprogress"
	"github.com/coreos/ioprogress"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
)

//https://developers.google.com/drive/api/v3/quickstart/go

// GoogleDriveProvider implements a provider using dropbox sdk
type GoogleDriveProvider struct {
	Config *oauth2.Config
	token  *oauth2.Token
}

func oauth2GoogleDriveConfig() *oauth2.Config {
	b := []byte(`{"installed":{"client_id":"26115953275-7971erj532s8d98vlso25467iudikbvf.apps.googleusercontent.com","project_id":"sharecmd-223413","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://www.googleapis.com/oauth2/v3/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"JblUhzPxWD-9zvJ7XBPr2Du8","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`)
	//config, err := google.ConfigFromJSON(b, drive.DriveMetadataScope)
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}

func NewGoogleDriveProvider(token string) *GoogleDriveProvider {
	tok := &oauth2.Token{}
	err := json.Unmarshal([]byte(token), tok)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err)
	}

	return &GoogleDriveProvider{token: tok, Config: oauth2GoogleDriveConfig()}
}

func (c *GoogleDriveProvider) getClient() *http.Client {
	return c.Config.Client(context.Background(), c.token)
}

func (c *GoogleDriveProvider) Upload(file *os.File, path string) (fileID string, err error) {

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	client := c.getClient()
	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	progressbar := &ioprogress.Reader{
		Reader: file,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: fileInfo.Size(),
	}
	parendID := getOrCreateFolder(srv, "sharecmd")

	f := &drive.File{
		Name:    filepath.Base(file.Name()),
		Parents: []string{parendID},
	}
	r, err := srv.Files.Create(f).Media(progressbar).Do()
	if err != nil {
		return "", err
	}
	return r.Id, nil
}

func (c *GoogleDriveProvider) GetLink(filepath string) (string, error) {
	fileID := filepath

	client := c.getClient()
	srv, err := drive.New(client)
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
	f, err := srv.Files.Get(fileID).Do()
	if err != nil {
		return "", err
	}
	link := fmt.Sprintf("https://drive.google.com/open?id=%s", f.Id)

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
	spew.Dump(r.Files)
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
