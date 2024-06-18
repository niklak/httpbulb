package httpbulb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{},
		Timeout:   10 * time.Second,
	}
}

func Test_Get(t *testing.T) {

	type serverResponse struct {
		URL     string      `json:"url"`
		Args    url.Values  `json:"args"`
		Headers http.Header `json:"headers"`
	}

	handleFunc := NewRouter()
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	apiURL, err := url.Parse(testServer.URL)
	require.NoError(t, err)
	apiURL.Path = "/get"
	apiURL.RawQuery = "k=v"

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	result := new(serverResponse)

	err = json.Unmarshal(body, result)

	require.NoError(t, err)

	require.Equal(t, apiURL.String(), result.URL)

	expectedArgs := url.Values{"k": []string{"v"}}

	require.Equal(t, expectedArgs, result.Args)

	parsedURL, err := url.Parse(result.URL)
	require.NoError(t, err)

	require.Equal(t, apiURL.Host, parsedURL.Host)

	log.Printf("%#v", result.URL)
}

func Test_HttpsGet(t *testing.T) {

	type serverResponse struct {
		URL string `json:"url"`
	}

	handleFunc := NewRouter()
	testServer := httptest.NewTLSServer(handleFunc)

	defer testServer.Close()

	testUrl := fmt.Sprintf("%s/get", testServer.URL)

	req, err := http.NewRequest("GET", testUrl, nil)
	require.NoError(t, err)

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	result := new(serverResponse)

	err = json.Unmarshal(body, result)

	require.NoError(t, err)
	require.True(t, strings.HasPrefix(result.URL, "https://"))
}

func Test_MethodNotAllowed(t *testing.T) {

	handleFunc := NewRouter()
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	testUrl := fmt.Sprintf("%s/get?k=v", testServer.URL)

	req, err := http.NewRequest("POST", testUrl, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func Test_Form(t *testing.T) {

	type serverResponse struct {
		URL  string     `json:"url"`
		Form url.Values `json:"form"`
	}

	handleFunc := NewRouter()
	// Start a test server that will act as a proxy
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	methods := []string{"post", "put", "patch"}

	for _, method := range methods {
		testURL := fmt.Sprintf("%s/%s", testServer.URL, method)

		req, err := http.NewRequest(strings.ToUpper(method), testURL, strings.NewReader("k=v"))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		// in this case we require either a result or a response
		result := new(serverResponse)

		json.Unmarshal(body, result)
		// ensure that result has the expected value
		require.Equal(t, testURL, result.URL)

		expectedForm := url.Values{"k": []string{"v"}}

		require.Equal(t, expectedForm, result.Form)
	}

}

func Test_JSON(t *testing.T) {
	type serverResponse struct {
		URL  string            `json:"url"`
		JSON map[string]string `json:"json"`
	}

	handleFunc := NewRouter()
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	methods := []string{"post", "put", "patch"}

	for _, method := range methods {
		testURL := fmt.Sprintf("%s/%s", testServer.URL, method)

		req, err := http.NewRequest(
			strings.ToUpper(method),
			testURL,
			strings.NewReader(`{"k":"v"}`),
		)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		// in this case we require either a result or a response
		result := new(serverResponse)

		json.Unmarshal(body, result)
		// ensure that result has the expected value
		require.Equal(t, testURL, result.URL)

		expectedJSON := map[string]string{"k": "v"}

		require.Equal(t, expectedJSON, result.JSON)
	}
}

func Test_PostMultipart(t *testing.T) {

	type serverResponse struct {
		URL   string              `json:"url"`
		Form  url.Values          `json:"form"`
		Files map[string][]string `json:"files"`
	}

	handleFunc := NewRouter()
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	testURL := fmt.Sprintf("%s/post", testServer.URL)

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	part, err := w.CreateFormFile("file", "file.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("file content"))
	require.NoError(t, err)
	err = w.WriteField("k", "v")
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req, err := http.NewRequest("POST", testURL, buf)
	require.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	result := new(serverResponse)

	err = json.Unmarshal(body, result)

	require.NoError(t, err)

	require.Equal(t, testURL, result.URL)

	expectedForm := url.Values{"k": []string{"v"}}

	require.Equal(t, expectedForm, result.Form)

	expectedFiles := map[string][]string{"file": {"file content"}}
	require.Equal(t, expectedFiles, result.Files)

}

func Test_Delete(t *testing.T) {

	type serverResponse struct {
		URL  string     `json:"url"`
		Args url.Values `json:"args"`
	}

	handleFunc := NewRouter()
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	testUrl := fmt.Sprintf("%s/delete?id=1", testServer.URL)

	req, err := http.NewRequest("DELETE", testUrl, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := new(serverResponse)
	json.Unmarshal(body, result)

	expectedArgs := url.Values{"id": []string{"1"}}

	require.Equal(t, expectedArgs, result.Args)
}

func Test_Anything(t *testing.T) {

	type serverResponse struct {
		URL string `json:"url"`
	}

	handleFunc := NewRouter()
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	testUrl := fmt.Sprintf("%s/anything?k=v", testServer.URL)

	req, err := http.NewRequest("GET", testUrl, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := new(serverResponse)
	json.Unmarshal(body, result)

	require.Equal(t, testUrl, result.URL)
}

func Test_AnythingAnything(t *testing.T) {
	type serverResponse struct {
		URL string `json:"url"`
	}

	handleFunc := NewRouter()
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	testUrl := fmt.Sprintf("%s/anything/something?k=v", testServer.URL)

	req, err := http.NewRequest("GET", testUrl, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := new(serverResponse)
	json.Unmarshal(body, result)
	require.Equal(t, testUrl, result.URL)
}
