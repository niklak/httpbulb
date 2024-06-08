package httpbulb

import (
	"compress/zlib"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	result := &serverResponse{}

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)
	assert.True(s.T(), result.Gzipped)

}

func (s *ResponseFormatSuite) TestDeflate() {

	// response will be automatically uncompressed by the http client,
	// since the transport does't specify `DisableCompression: true`
	type serverResponse struct {
		Deflated bool `json:"deflated"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/deflate", nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	result := &serverResponse{}

	reader, err := zlib.NewReader(resp.Body)
	assert.NoError(s.T(), err)

	body, err := io.ReadAll(reader)
	assert.NoError(s.T(), err)
	reader.Close()

	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)
	assert.True(s.T(), result.Deflated)

}

func (s *ResponseFormatSuite) TestBrotli() {

	// response will be automatically uncompressed by the http client,
	// since the transport does't specify `DisableCompression: true`
	type serverResponse struct {
		Brotli bool `json:"brotli"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/brotli", nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	result := &serverResponse{}

	reader := brotli.NewReader(resp.Body)

	body, err := io.ReadAll(reader)
	assert.NoError(s.T(), err)

	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)
	assert.True(s.T(), result.Brotli)

}

func (s *ResponseFormatSuite) TestRobots() {

	req, err := http.NewRequest("GET", s.testServer.URL+"/robots.txt", nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	expectedBody := []byte(`
User-agent: *
Disallow: /deny
	`)

	assert.Equal(s.T(), expectedBody, body)
}

func TestResponseFormatSuite(t *testing.T) {
	suite.Run(t, new(ResponseFormatSuite))
}
