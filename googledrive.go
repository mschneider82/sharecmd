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
	"github.com/mitchellh/ioprogress"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v2"
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
	spew.Dump(tok)

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
	p := &drive.ParentReference{Id: parendID}
	permission := &drive.Permission{
		Type:     "anyone",
		Role:     "reader",
		WithLink: true,
	}
	f := &drive.File{
		Title:          filepath.Base(file.Name()),
		Parents:        []*drive.ParentReference{p},
		Permissions:    []*drive.Permission{permission},
		UserPermission: permission,
		Shared:         true,
	}
	r, err := srv.Files.Insert(f).Media(progressbar).Do()

	return r.Id, err
}

func (c *GoogleDriveProvider) GetLink(filepath string) (string, error) {
	fileID := filepath

	client := c.getClient()
	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	permission := &drive.Permission{
		Type:     "anyone",
		Role:     "reader",
		WithLink: true,
	}
	f := &drive.File{
		Id:             fileID,
		Permissions:    []*drive.Permission{permission},
		UserPermission: permission,
		Shared:         true,
	}

	r, err := srv.Files.Update(fileID, f).Do()
	spew.Dump(r)
	spew.Dump(err)

	return "", nil
}

func getOrCreateFolder(d *drive.Service, folderName string) string {
	folderID := ""
	if folderName == "" {
		folderName = "sharecmd"
	}
	q := fmt.Sprintf("title=\"%s\" and mimeType=\"application/vnd.google-apps.folder\"", folderName)

	r, err := d.Files.List().Q(q).MaxResults(1).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve foldername.", err)
	}

	if len(r.Items) > 0 {
		folderID = r.Items[0].Id
	} else {
		// no folder found create new
		fmt.Printf("Folder not found. Create new folder : %s\n", folderName)
		f := &drive.File{Title: folderName, Description: "Auto Create by sharecmd", MimeType: "application/vnd.google-apps.folder"}
		r, err := d.Files.Insert(f).Do()
		if err != nil {
			fmt.Printf("An error occurred when create folder: %v\n", err)
		}
		folderID = r.Id
	}
	return folderID
}
