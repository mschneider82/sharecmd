package nextcloud

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos/ioprogress"
	"github.com/davecgh/go-spew/spew"
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

func NewProvider(c Config) (*Provider, error) {
	return &Provider{}, nil
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

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/remote.php/dav/sharecmd/%s", s.config.URL, filename), progressbar)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.config.Username, s.config.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return "", nil
}

func (s *Provider) GetLink(filename string) (string, error) {
	body := strings.NewReader(fmt.Sprintf(`path=sharecmd/%s&shareType=3&permissions=1`, filename))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/ocs/v1.php/apps/files_sharing/api/v1/shares", s.config.URL), body)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.config.Username, s.config.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	return string(b), nil
}

func (s *Provider) createFolder(foldername string) error {
	req, err := http.NewRequest("MKCOL", fmt.Sprintf("%s/nextcloud/remote.php/dav/files/%s/%s", s.config.URL, s.config.Username, foldername), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.config.Username, s.config.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	spew.Dump(resp.Body)
}

type Reply struct {
	Exception string `xml:"exception"`
	Message   string `xml:"message"`
}
