package biturl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	resp, err := http.Post(fmt.Sprintf("https://api.biturl.top/short?url=%s", b.URL), "", nil)
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
