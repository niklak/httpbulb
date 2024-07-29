package httpbulb

import (
	"bytes"
	"crypto/tls"
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

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{},
		Timeout:   10 * time.Second,
	}
}

type MethodsSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *MethodsSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewUnstartedServer(handleFunc)
	s.testServer.EnableHTTP2 = true
	s.testServer.StartTLS()

	s.client = s.testServer.Client()
	s.client.Timeout = 10 * time.Second
}

func (s *MethodsSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *MethodsSuite) TestGet() {
	type serverResponse struct {
		URL     string      `json:"url"`
		Args    url.Values  `json:"args"`
		Headers http.Header `json:"headers"`
	}

	t := s.T()
	apiURL, err := url.Parse(s.testServer.URL)
	require.NoError(t, err)
	apiURL.Path = "/get"
	apiURL.RawQuery = "k=v"

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	require.NoError(t, err)

	resp, err := s.client.Do(req)
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

	require.Equal(t, apiURL.Host, result.Headers.Get("Host"))

}
func (s *MethodsSuite) TestHttp2Client() {
	type testArgs struct {
		name          string
		client        *http.Client
		wantProto     string
		wantClientErr bool
	}

	type serverResponse struct {
		URL   string `json:"url"`
		Proto string `json:"proto"`
	}

	testUrl := fmt.Sprintf("%s/get", s.testServer.URL)

	h2Client := s.testServer.Client()

	h1Client := &http.Client{}

	h1ClientInsecure := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	forcedH2Client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, ForceAttemptHTTP2: true},
	}

	tests := []testArgs{
		{name: "Forced secure HTTP/2.0", client: h2Client, wantProto: "HTTP/2.0"},
		{name: "Forced insecure HTTP/2.0", client: forcedH2Client, wantProto: "HTTP/2.0"},
		{name: "HTTP/1.1 - Insecure", client: h1ClientInsecure, wantProto: "HTTP/1.1"},
		{name: "HTTP/1.1 - Without CA Certs", client: h1Client, wantClientErr: true},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", testUrl, nil)
			require.NoError(t, err)
			resp, err := tt.client.Do(req)
			if tt.wantClientErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			result := new(serverResponse)

			err = json.Unmarshal(body, result)

			require.NoError(t, err)
			require.True(t, strings.HasPrefix(result.URL, "https://"))

			require.Equal(t, tt.wantProto, result.Proto)
		})
	}
}

func (s *MethodsSuite) TestMethodsInternalError() {

	type testArgs struct {
		name        string
		apiURL      string
		method      string
		contentType string
		body        []byte
	}

	tests := []testArgs{
		{
			name:        "bad form",
			apiURL:      fmt.Sprintf("%s/post", s.testServer.URL),
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			body:        []byte("%"),
		},
		{
			name:        "bad multipart form",
			apiURL:      fmt.Sprintf("%s/post", s.testServer.URL),
			method:      http.MethodPost,
			contentType: "multipart/form-data",
			body:        []byte("%"),
		},
		{
			name:        "bad json",
			apiURL:      fmt.Sprintf("%s/post", s.testServer.URL),
			method:      http.MethodPost,
			contentType: "application/json",
			body:        []byte("%"),
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest(tt.method, tt.apiURL, bytes.NewReader(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tt.contentType)

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		})
	}
}

func (s *MethodsSuite) TestMethodNotAllowed() {

	testUrl := fmt.Sprintf("%s/get?k=v", s.testServer.URL)

	req, err := http.NewRequest("POST", testUrl, nil)
	require.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	require.NoError(s.T(), err)

	require.Equal(s.T(), http.StatusMethodNotAllowed, resp.StatusCode)
}

func (s *MethodsSuite) TestForm() {

	type serverResponse struct {
		URL  string     `json:"url"`
		Form url.Values `json:"form"`
	}

	type testArgs struct {
		method string
	}

	tests := []testArgs{
		{"post"}, {"put"}, {"patch"},
	}

	for _, tt := range tests {
		s.T().Run(tt.method, func(t *testing.T) {
			testURL := fmt.Sprintf("%s/%s", s.testServer.URL, tt.method)

			req, err := http.NewRequest(strings.ToUpper(tt.method), testURL, strings.NewReader("k=v"))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := s.client.Do(req)
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
		})

	}

}

func (s *MethodsSuite) TestJSON() {
	type serverResponse struct {
		URL  string            `json:"url"`
		JSON map[string]string `json:"json"`
	}

	type testArgs struct {
		method string
	}

	tests := []testArgs{{"post"}, {"put"}, {"patch"}}

	for _, tt := range tests {
		s.T().Run(tt.method, func(t *testing.T) {
			testURL := fmt.Sprintf("%s/%s", s.testServer.URL, tt.method)

			req, err := http.NewRequest(
				strings.ToUpper(tt.method),
				testURL,
				strings.NewReader(`{"k":"v"}`),
			)
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := s.client.Do(req)
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
		})

	}
}

func (s *MethodsSuite) TestPostMultipart() {

	type serverResponse struct {
		URL   string              `json:"url"`
		Form  url.Values          `json:"form"`
		Files map[string][]string `json:"files"`
	}

	testURL := fmt.Sprintf("%s/post", s.testServer.URL)

	t := s.T()

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

	resp, err := s.client.Do(req)
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

func (s *MethodsSuite) TestDelete() {

	type serverResponse struct {
		URL  string     `json:"url"`
		Args url.Values `json:"args"`
	}

	t := s.T()

	testUrl := fmt.Sprintf("%s/delete?id=1", s.testServer.URL)

	req, err := http.NewRequest("DELETE", testUrl, nil)
	require.NoError(t, err)

	resp, err := s.client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := new(serverResponse)
	json.Unmarshal(body, result)

	expectedArgs := url.Values{"id": []string{"1"}}

	require.Equal(t, expectedArgs, result.Args)
}

func (s *MethodsSuite) TestAnything() {

	type serverResponse struct {
		URL string `json:"url"`
	}

	type testArgs struct {
		name string
		url  string
	}

	tests := []testArgs{
		{name: "anything", url: fmt.Sprintf("%s/anything?k=v", s.testServer.URL)},
		{name: "anything path", url: fmt.Sprintf("%s/anything/something?k=v", s.testServer.URL)},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			result := new(serverResponse)
			json.Unmarshal(body, result)

			require.Equal(t, tt.url, result.URL)
		})
	}

}

func TestMethodsSuite(t *testing.T) {
	suite.Run(t, new(MethodsSuite))
}
