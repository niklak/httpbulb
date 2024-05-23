package httpbulb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RequestInspectionSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *RequestInspectionSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *RequestInspectionSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *RequestInspectionSuite) TestHeaders() {

	type serverResponse struct {
		Headers http.Header `json:"headers"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/headers", nil)
	assert.NoError(s.T(), err)

	req.Header.Set("X-Test-Header", "test")

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	assert.NoError(s.T(), err)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), "test", result.Headers.Get("X-Test-Header"))

}

func (s *RequestInspectionSuite) TestUserAgent() {

	type serverResponse struct {
		UserAgent string `json:"user-agent"`
	}

	userAgent := "bulb/0.1"

	req, err := http.NewRequest("GET", s.testServer.URL+"/user-agent", nil)
	assert.NoError(s.T(), err)

	req.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	assert.NoError(s.T(), err)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), userAgent, result.UserAgent)
}

func (s *RequestInspectionSuite) TestIP() {

	type serverResponse struct {
		Origin string `json:"origin"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/ip", nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	assert.NoError(s.T(), err)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)

	assert.True(s.T(), strings.HasPrefix(result.Origin, "127.0.0.1"))
}

func TestRequestInspectionSuite(t *testing.T) {
	suite.Run(t, new(RequestInspectionSuite))
}
