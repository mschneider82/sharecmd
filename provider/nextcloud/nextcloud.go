package nextcloud

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos/ioprogress"
	humanize "github.com/dustin/go-humanize"
)

type Config struct {
	URL      string
	Username string
	Password string
}

type Provider struct {
	config Config
}

func NewProvider(c Config) *Provider {
	return &Provider{config: c}
}

func (s *Provider) Upload(file *os.File, path string) (string, error) {
	if err := s.createFolder("sharecmd"); err != nil {
		fmt.Printf("could not create folder: %s\n", err.Error())
	}

	filename := filepath.Base(file.Name())
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	progressbar := &ioprogress.Reader{
		Reader: file,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: fileInfo.Size(),
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/remote.php/webdav/sharecmd/%s", s.config.URL, filename), progressbar)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.config.Username, s.config.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("OCS-APIRequest", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return filename, nil
}

func (s *Provider) GetLink(filename string) (string, error) {
	body := strings.NewReader(fmt.Sprintf(`path=sharecmd/%s&shareType=3&permissions=1`, filename))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/ocs/v1.php/apps/files_sharing/api/v1/shares", s.config.URL), body)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.config.Username, s.config.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("OCS-APIRequest", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	var reply struct {
		XMLName xml.Name `xml:"ocs"`
		Text    string   `xml:",chardata"`
		Meta    struct {
			Text         string `xml:",chardata"`
			Status       string `xml:"status"`
			Statuscode   string `xml:"statuscode"`
			Message      string `xml:"message"`
			Totalitems   string `xml:"totalitems"`
			Itemsperpage string `xml:"itemsperpage"`
		} `xml:"meta"`
		Data struct {
			Text                 string `xml:",chardata"`
			ID                   string `xml:"id"`
			ShareType            string `xml:"share_type"`
			UidOwner             string `xml:"uid_owner"`
			DisplaynameOwner     string `xml:"displayname_owner"`
			Permissions          string `xml:"permissions"`
			Stime                string `xml:"stime"`
			Parent               string `xml:"parent"`
			Expiration           string `xml:"expiration"`
			Token                string `xml:"token"`
			UidFileOwner         string `xml:"uid_file_owner"`
			Note                 string `xml:"note"`
			DisplaynameFileOwner string `xml:"displayname_file_owner"`
			Path                 string `xml:"path"`
			ItemType             string `xml:"item_type"`
			Mimetype             string `xml:"mimetype"`
			StorageID            string `xml:"storage_id"`
			Storage              string `xml:"storage"`
			ItemSource           string `xml:"item_source"`
			FileSource           string `xml:"file_source"`
			FileParent           string `xml:"file_parent"`
			FileTarget           string `xml:"file_target"`
			ShareWith            string `xml:"share_with"`
			ShareWithDisplayname string `xml:"share_with_displayname"`
			URL                  string `xml:"url"`
			MailSend             string `xml:"mail_send"`
		} `xml:"data"`
	}

	err = xml.Unmarshal(b, &reply)
	if err != nil {
		return "", err
	}
	if reply.Data.URL == "" {
		return "", fmt.Errorf("Status: %s, Message: %s ", reply.Meta.Status, reply.Meta.Message)
	}
	return reply.Data.URL, nil
}

func (s *Provider) createFolder(foldername string) error {
	url := fmt.Sprintf("%s/remote.php/dav/files/%s/%s", s.config.URL, s.config.Username, foldername)

	req, err := http.NewRequest("MKCOL", url, nil)
	if err != nil {
		return err
	}
	h := make(map[string][]string)
	h["OCS-APIRequest"] = []string{"true"}
	req.Header = h
	req.SetBasicAuth(s.config.Username, s.config.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
