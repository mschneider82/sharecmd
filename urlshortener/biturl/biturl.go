package biturl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// BitURL Shortener Interface impl
type BitURL struct {
	URL string
}

// New creates a new BitURL
func New(url string) *BitURL {
	return &BitURL{URL: url}
}

// SetupQuestions will ask for user api key or something else
func (b *BitURL) SetupQuestions() map[string]string {
	// BitURL has no api key so new questions!
	return make(map[string]string)
}

// ShortURL maks http post to biturl.top to get short url
func (b *BitURL) GetName() string {
	return "biturl"
}

// ShortURL maks http post to biturl.top to get short url
func (b *BitURL) ShortURL() (string, error) {
	body := `-----------------------------4139599310243699752647233313
Content-Disposition: form-data; name="url"

` + b.URL + `
-----------------------------4139599310243699752647233313--`
	resp, err := http.Post("https://api.biturl.top/short", "multipart/form-data; boundary=---------------------------4139599310243699752647233313", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	resultBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expecting json got: %s", err.Error(), string(resultBody))
	}
	// {"result":true,"short":"https://biturl.top/EbQjye","message":""}
	var reply struct {
		Result  bool   `json:"result"`
		Short   string `json:"short"`
		Message string `json:"message"`
	}

	err = json.Unmarshal(resultBody, &reply)
	if err != nil {
		return "", fmt.Errorf("result body error: %s, expecting json got: %s", err.Error(), string(resultBody))
	}

	if len(reply.Short) > 0 {
		return reply.Short, nil
	}
	return "", fmt.Errorf("%s", string(resultBody))
}
