package httpbulb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestResponseFormatSuite(t *testing.T) {
	suite.Run(t, new(ResponseFormatSuite))
}
