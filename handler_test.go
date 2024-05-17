package httpbulb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Get(t *testing.T) {

	type httpBinResponse struct {
		URL   string     `json:"url"`
		Args  url.Values `json:"args"`
		Error string     `json:"error"`
	}

	handleFunc := NewRouter()
	// Start a test server that will act as a proxy
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	// Create an http client with a proxy

	httpClient := &http.Client{
		Transport: &http.Transport{},
		Timeout:   10 * time.Second,
	}

	testGETUrl := fmt.Sprintf("%s/get?k=v", testServer.URL)

	req, err := http.NewRequest("GET", testGETUrl, nil)
	assert.NoError(t, err)

	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	// in this case we require either a result or a response
	result := new(httpBinResponse)

	err = json.Unmarshal(body, result)

	assert.NoError(t, err)

	assert.Equal(t, "", result.Error)

	// ensure that result has the expected value
	assert.Equal(t, testGETUrl, result.URL)

	expectedArgs := url.Values{"k": []string{"v"}}

	assert.Equal(t, expectedArgs, result.Args)
}

func Test_PostForm(t *testing.T) {

	type httpBinResponse struct {
		URL   string     `json:"url"`
		Form  url.Values `json:"form"`
		Error string     `json:"error"`
	}

	handleFunc := NewRouter()
	// Start a test server that will act as a proxy
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	// Create an http client with a proxy

	httpClient := &http.Client{
		Transport: &http.Transport{},
		Timeout:   10 * time.Second,
	}

	testURL := fmt.Sprintf("%s/post", testServer.URL)

	req, err := http.NewRequest("POST", testURL, strings.NewReader("k=v"))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	// in this case we require either a result or a response
	result := new(httpBinResponse)

	err = json.Unmarshal(body, result)

	assert.NoError(t, err)

	assert.Equal(t, "", result.Error)

	// ensure that result has the expected value
	assert.Equal(t, testURL, result.URL)

	expectedForm := url.Values{"k": []string{"v"}}

	assert.Equal(t, expectedForm, result.Form)

}

func Test_PostMultipart(t *testing.T) {

	type httpBinResponse struct {
		URL   string              `json:"url"`
		Form  url.Values          `json:"form"`
		Files map[string][]string `json:"files"`
		Error string              `json:"error"`
	}

	handleFunc := NewRouter()
	// Start a test server that will act as a proxy
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	// Create an http client with a proxy

	httpClient := &http.Client{
		Transport: &http.Transport{},
		Timeout:   10 * time.Second,
	}

	testURL := fmt.Sprintf("%s/post", testServer.URL)

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	part, err := w.CreateFormFile("file", "file.txt")
	assert.NoError(t, err)

	_, err = part.Write([]byte("file content"))
	assert.NoError(t, err)

	err = w.WriteField("k", "v")
	assert.NoError(t, err)

	assert.NoError(t, w.Close())

	req, err := http.NewRequest("POST", testURL, buf)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	// in this case we require either a result or a response
	result := new(httpBinResponse)

	err = json.Unmarshal(body, result)

	assert.NoError(t, err)

	assert.Equal(t, "", result.Error)

	// ensure that result has the expected value
	assert.Equal(t, testURL, result.URL)

	expectedForm := url.Values{"k": []string{"v"}}

	assert.Equal(t, expectedForm, result.Form)

	expectedFiles := map[string][]string{"file": {"file content"}}
	assert.Equal(t, expectedFiles, result.Files)

}
