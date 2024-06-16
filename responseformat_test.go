package httpbulb

import (
	"compress/zlib"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ResponseFormatSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *ResponseFormatSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *ResponseFormatSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *ResponseFormatSuite) TestGzip() {

	// response will be automatically uncompressed by the http client,
	// since the transport does't specify `DisableCompression: true`
	type serverResponse struct {
		Gzipped bool `json:"gzipped"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/gzip", nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	result := &serverResponse{}

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	err = json.Unmarshal(body, result)
	s.Require().NoError(err)
	s.Require().True(result.Gzipped)

}

func (s *ResponseFormatSuite) TestDeflate() {

	// response will be automatically uncompressed by the http client,
	// since the transport does't specify `DisableCompression: true`
	type serverResponse struct {
		Deflated bool `json:"deflated"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/deflate", nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	result := &serverResponse{}

	reader, err := zlib.NewReader(resp.Body)
	s.Require().NoError(err)

	body, err := io.ReadAll(reader)
	s.Require().NoError(err)
	reader.Close()

	err = json.Unmarshal(body, result)
	s.Require().NoError(err)
	s.Require().True(result.Deflated)

}

func (s *ResponseFormatSuite) TestBrotli() {

	// response will be automatically uncompressed by the http client,
	// since the transport does't specify `DisableCompression: true`
	type serverResponse struct {
		Brotli bool `json:"brotli"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/brotli", nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	result := &serverResponse{}

	reader := brotli.NewReader(resp.Body)

	body, err := io.ReadAll(reader)
	s.Require().NoError(err)

	err = json.Unmarshal(body, result)
	s.Require().NoError(err)
	s.Require().True(result.Brotli)

}

func (s *ResponseFormatSuite) TestRobots() {

	req, err := http.NewRequest("GET", s.testServer.URL+"/robots.txt", nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	expectedBody := []byte(`
User-agent: *
Disallow: /deny
	`)

	s.Require().Equal(expectedBody, body)
}

func (s *ResponseFormatSuite) TestDeny() {

	req, err := http.NewRequest("GET", s.testServer.URL+"/deny", nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	s.Require().Contains(string(body), "YOU SHOULDN'T BE HERE\n")
}

func (s *ResponseFormatSuite) TestSamples() {
	type testArgs struct {
		name             string
		apiPath          string
		wantContentType  string
		wantStatusCode   int
		wantBodyFragment string
	}

	tests := []testArgs{
		{
			name:             "encoding/utf8",
			apiPath:          "/encoding/utf8",
			wantContentType:  "text/html; charset=utf-8",
			wantStatusCode:   http.StatusOK,
			wantBodyFragment: `ᚻᛖ ᚳᚹᚫᚦ ᚦᚫᛏ ᚻᛖ ᛒᚢᛞᛖ ᚩᚾ ᚦᚫᛗ ᛚᚪᚾᛞᛖ ᚾᚩᚱᚦᚹᛖᚪᚱᛞᚢᛗ ᚹᛁᚦ ᚦᚪ ᚹᛖᛥᚫ`,
		},
		{
			name:             "html",
			apiPath:          "/html",
			wantContentType:  "text/html; charset=utf-8",
			wantStatusCode:   http.StatusOK,
			wantBodyFragment: `<h1>Herman Melville - Moby-Dick</h1>`,
		},
		{
			name:             "json",
			apiPath:          "/json",
			wantContentType:  "application/json",
			wantStatusCode:   http.StatusOK,
			wantBodyFragment: `"title": "Sample Slide Show"`,
		},
		{
			name:             "xml",
			apiPath:          "/xml",
			wantContentType:  "application/xml",
			wantStatusCode:   http.StatusOK,
			wantBodyFragment: `title="Sample Slide Show"`,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", s.testServer.URL+tt.apiPath, nil)
			require.NoError(t, err)

			resp, err := s.client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)
			require.Equal(t, tt.wantContentType, resp.Header.Get("Content-Type"))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Contains(t, string(body), tt.wantBodyFragment)
		})

	}
}

func TestResponseFormatSuite(t *testing.T) {
	suite.Run(t, new(ResponseFormatSuite))
}
