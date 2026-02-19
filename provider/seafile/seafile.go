package seafile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/mschneider82/easygo"
)

type Config struct {
	URL              string
	Username         string
	Password         string
	TwoFactorEnabled bool
	OTP              string
	RepoID           string
}

// GetToken from seafile
func (c *Config) GetToken() (string, error) {
	body := strings.NewReader(fmt.Sprintf(`username=%s&password=%s`, c.Username, c.Password))
	req, err := http.NewRequest("POST", c.URL+"/api2/auth-token/", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.TwoFactorEnabled {
		req.Header.Set("X-Seafile-Otp", c.OTP)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		Token string `json:"token"`
	}
	resultBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expting token got: %s", err.Error(), string(resultBody))
	}
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err.Error())
	}

	return response.Token, nil
}

func (c *Config) CreateLibrary(token string) error {
	body := strings.NewReader(`name=sharecmd&desc=ShareCmd`)
	req, err := http.NewRequest("POST", c.URL+"/api2/repos/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
	req.Header.Set("Accept", "application/json; indent=4")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("Library sharecmd created.")
	}

	var response struct {
		RepoID string `json:"repo_id"`
	}
	resultBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("result body error: %s, expecting repoid got: %s", err.Error(), string(resultBody))
	}
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return fmt.Errorf("json unmarshal error: %s", err.Error())
	}

	c.RepoID = response.RepoID

	return nil
}

// Provider ..
type Provider struct {
	URL    string
	Token  string
	RepoID string
}

func (s *Provider) Upload(r io.Reader, filename string, size int64) (fileID string, err error) {
	// get upload link
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api2/repos/%s/upload-link/?p=/&replace=1", s.URL, s.RepoID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", s.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	uploadLinkBroken, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expecting repoid got: %s", err.Error(), string(uploadLinkBroken))
	}
	uploadLink := easygo.StringStrip(string(uploadLinkBroken), `"`)

	_, err = uploadfile(uploadLink, "/", filename, s.Token, r)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func uploadfile(uploadlink, folder, filename, token string, src io.Reader) (string, error) {
	requestbody := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(requestbody)
	part, err := multipartWriter.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, src)

	multipartWriter.WriteField("filename", filename)
	multipartWriter.WriteField("parent_dir", folder)

	err = multipartWriter.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", uploadlink, requestbody)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Token "+token)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	responsebody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return string(responsebody), nil
}

func (s *Provider) GetLink(filepath string) (string, error) {
	body := strings.NewReader(fmt.Sprintf(`p=/%s`, filepath))
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api2/repos/%s/file/shared-link/", s.URL, s.RepoID), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", s.Token))
	req.Header.Set("Accept", "application/json; indent=4")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if len(resp.Header.Get("Location")) == 0 {
		return "", errors.New("expecting location header from seafile")
	}
	url := fmt.Sprintf("%s?dl=1", resp.Header.Get("Location"))
	return url, nil
}

func NewProvider(url, token, repoid string) *Provider {
	return &Provider{URL: url, Token: token, RepoID: repoid}
}
