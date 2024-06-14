package httpbulb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ResponseInspectionSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *ResponseInspectionSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *ResponseInspectionSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *ResponseInspectionSuite) TestResponseHeaders() {

	apiURL, err := url.Parse(s.testServer.URL)
	assert.NoError(s.T(), err)
	apiURL.Path = "/response-headers"

	query := url.Values{}
	query.Add("x-test-header", "1")
	query.Add("X-Test-Header", "2")
	query.Add("X-Test-Header", "3")
	apiURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)
	resp.Body.Close()

	result := http.Header{}

	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	expectedHeaderValue := []string{"1", "2", "3"}
	assert.Subset(s.T(), expectedHeaderValue, result["X-Test-Header"])
	assert.Len(s.T(), result["X-Test-Header"], len(expectedHeaderValue))

}

func TestResponseInspectionSuite(t *testing.T) {
	suite.Run(t, new(ResponseInspectionSuite))
}
