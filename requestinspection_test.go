package httpbulb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	s.Require().NoError(err)

	req.Header.Set("X-Test-Header", "test")

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	s.Require().NoError(err)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	s.Require().NoError(err)

	s.Require().Equal("test", result.Headers.Get("X-Test-Header"))

}

func (s *RequestInspectionSuite) TestUserAgent() {

	type serverResponse struct {
		UserAgent string `json:"user-agent"`
	}

	userAgent := "bulb/0.1"

	req, err := http.NewRequest("GET", s.testServer.URL+"/user-agent", nil)
	s.Require().NoError(err)

	req.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	s.Require().NoError(err)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	s.Require().NoError(err)

	s.Require().Equal(userAgent, result.UserAgent)
}

func (s *RequestInspectionSuite) TestIP() {

	type serverResponse struct {
		Origin string `json:"origin"`
	}

	req, err := http.NewRequest("GET", s.testServer.URL+"/ip", nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	s.Require().NoError(err)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	s.Require().NoError(err)

	s.Require().True(strings.HasPrefix(result.Origin, "127.0.0.1"))
}

func TestRequestInspectionSuite(t *testing.T) {
	suite.Run(t, new(RequestInspectionSuite))
}
