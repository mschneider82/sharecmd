package opendrive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/coreos/ioprogress"
	"github.com/dustin/go-humanize"
)

// NewProvider creates a new Provider
func NewProvider(user, pass string) *Provider {
	return &Provider{Username: user, Passwd: pass}
}

type Provider struct {
	Username     string `json:"username"`
	Passwd       string `json:"passwd"`
	downloadlink string
}

func (o *Provider) getSessionID() (string, error) {
	body, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("https://dev.opendrive.com/api/v1/session/login.json", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		SessionID string `json:"SessionID"`
	}
	resultBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expting sessionid got: %s", err.Error(), string(resultBody))
	}
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err.Error())
	}

	return response.SessionID, nil
}

func (o *Provider) createFolder(sessionid string) (string, error) {
	type Props struct {
		SessionID  string `json:"session_id"`
		FolderName string `json:"folder_name"`
	}

	props := Props{
		SessionID:  sessionid,
		FolderName: "sharecmd",
	}

	body, err := json.Marshal(props)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("https://dev.opendrive.com/api/v1/folder.json", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		FolderID string `json:"FolderID"`
	}
	resultBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expting sessionid got: %s", err.Error(), string(resultBody))
	}
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err.Error())
	}

	return response.FolderID, nil
}

func (o *Provider) getFolderID(sessionid string) (string, error) {
	type Props struct {
		SessionID string `json:"session_id"`
		Path      string `json:"path"`
	}

	props := Props{
		SessionID: sessionid,
		Path:      "sharecmd",
	}

	body, err := json.Marshal(props)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("https://dev.opendrive.com/api/v1/folder/idbypath.json", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("folder not found")
	}
	var response struct {
		FolderID string `json:"FolderId"`
	}
	resultBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expting sessionid got: %s", err.Error(), string(resultBody))
	}
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err.Error())
	}
	return response.FolderID, nil
}

// createFile returns the fileId if sucessful
func (o *Provider) createFile(sessionid, folderid, filename string) (fileid string, downloadlink string, err error) {
	type Props struct {
		SessionID    string `json:"session_id"`
		FolderID     string `json:"folder_id"`
		Filename     string `json:"file_name"`
		OpenIfExists int    `json:"open_if_exists"`
	}

	props := Props{
		SessionID:    sessionid,
		FolderID:     folderid,
		Filename:     filename,
		OpenIfExists: 1,
	}

	body, err := json.Marshal(props)
	if err != nil {
		return "", "", err
	}
	resp, err := http.Post("https://dev.opendrive.com/api/v1/upload/create_file.json", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", "", fmt.Errorf("folder not found")
	}
	var response struct {
		FileId       string `json:"FileId"`
		DownloadLink string `json:"DownloadLink"`
	}
	resultBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("result body error: %s, expting sessionid got: %s", err.Error(), string(resultBody))
	}
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return "", "", fmt.Errorf("json unmarshal error: %s", err.Error())
	}

	return response.FileId, response.DownloadLink, nil
}

func (o *Provider) openfileUpload(sessionid, fileID, fileName string, src *os.File) (string, error) {
	type Props struct {
		SessionID    string `json:"session_id"`
		FileID       string `json:"file_id"`
		FileSize     int    `json:"file_size"`
		TempLocation string `json:"temp_location,omitempty"`
	}
	finfo, err := src.Stat()
	if err != nil {
		return "", err
	}

	props := Props{
		SessionID: sessionid,
		FileID:    fileID,
		FileSize:  int(finfo.Size()),
	}
	body, err := json.Marshal(props)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("https://dev.opendrive.com/api/v1/upload/open_file_upload.json", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		TempLocation       string `json:"TempLocation"`
		RequireCompression bool   `json:"RequireCompression"`
		RequireHash        bool   `json:"RequireHash"`
		RequireHashOnly    bool   `json:"RequireHashOnly"`
	}
	resultBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expting sessionid got: %s", err.Error(), string(resultBody))
	}
	if resp.StatusCode == 403 {
		return "", fmt.Errorf("%s", resultBody)
	}

	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err.Error())
	}

	body2 := &bytes.Buffer{}
	writer := multipart.NewWriter(body2)
	writer.WriteField("session_id", sessionid)
	writer.WriteField("file_id", fileID)
	writer.WriteField("temp_location", response.TempLocation)
	writer.WriteField("chunk_offset", "0")
	writer.WriteField("chunk_size", fmt.Sprintf("%d", finfo.Size()))
	w, err := writer.CreateFormFile("file_data", fileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(w, src)
	if err != nil {
		return "", err
	}
	src.Close()
	err = writer.Close()
	if err != nil {
		return "", err
	}
	progressbar := &ioprogress.Reader{
		Reader: body2,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: int64(body2.Len()),
	}

	resp2, err := http.Post("https://dev.opendrive.com/api/v1/upload/upload_file_chunk.json", writer.FormDataContentType(), progressbar)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()
	props.TempLocation = response.TempLocation
	body, _ = json.Marshal(props)

	resp3, err := http.Post("https://dev.opendrive.com/api/v1/upload/close_file_upload.json", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp3.Body.Close()

	resultBody3, err := ioutil.ReadAll(resp3.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expting sessionid got: %s", err.Error(), string(resultBody))
	}

	var response3 struct {
		DownloadLink string `json:"DownloadLink"`
	}
	err = json.Unmarshal(resultBody3, &response3)
	if err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err.Error())
	}
	if response3.DownloadLink == "" {
		resultBody2, _ := ioutil.ReadAll(resp2.Body)
		return "", fmt.Errorf("no downloadlink got: %s\n %s\n %s", string(resultBody), string(resultBody2), string(resultBody3))
	}
	return response3.DownloadLink, nil
}

func (o *Provider) getOrCreateFolderID(sessionid string) (string, error) {
	fid, err := o.getFolderID(sessionid)
	if err != nil {
		fid, err = o.createFolder(sessionid)
		if err != nil {
			return "", err
		}
		return fid, nil
	}
	return fid, nil
}

func (o *Provider) Upload(file *os.File, path string) (string, error) {
	sid, err := o.getSessionID()
	if err != nil {
		return "", err
	}
	folderID, err := o.getOrCreateFolderID(sid)
	if err != nil {
		return "", err
	}

	filename := filepath.Base(file.Name())
	fileid, _, err := o.createFile(sid, folderID, filename)
	if err != nil {
		return "", err
	}

	downloadlink, err := o.openfileUpload(sid, fileid, filename, file)
	if err != nil {
		return "", err
	}
	o.downloadlink = downloadlink
	return downloadlink, nil
}

func (o *Provider) GetLink(string) (string, error) {
	if len(o.downloadlink) == 0 {
		return "", fmt.Errorf("failure")
	}
	return o.downloadlink, nil
}
