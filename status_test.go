package httpbulb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StatusCodeSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *StatusCodeSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *StatusCodeSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *StatusCodeSuite) TestGetOK() {

	resp, err := s.requestStatusCode("GET", http.StatusOK)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StatusCodeSuite) TestPostOK() {

	resp, err := s.requestStatusCode("POST", http.StatusOK)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StatusCodeSuite) TestPutOK() {

	resp, err := s.requestStatusCode("PUT", http.StatusOK)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StatusCodeSuite) TestPatchOK() {

	resp, err := s.requestStatusCode("PATCH", http.StatusOK)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StatusCodeSuite) TestDeleteOK() {

	resp, err := s.requestStatusCode("DELETE", http.StatusOK)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StatusCodeSuite) TestBadMethod() {

	resp, err := s.requestStatusCode("Get", http.StatusOK)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusMethodNotAllowed, resp.StatusCode)
}

func (s *StatusCodeSuite) TestBadStatusCode() {
	// not found because status code is not matching regex
	resp, err := s.requestStatusCode("GET", 667)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *StatusCodeSuite) TestCustomCode() {
	resp, err := s.requestStatusCode("GET", 444)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 444, resp.StatusCode)

}

func (s *StatusCodeSuite) requestStatusCode(method string, code int) (*http.Response, error) {
	testURL := fmt.Sprintf("%s/status/%d", s.testServer.URL, code)

	req, err := http.NewRequest(method, testURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return resp, nil
}

func TestStatusCodeSuite(t *testing.T) {
	suite.Run(t, new(StatusCodeSuite))
}
