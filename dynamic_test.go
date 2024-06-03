package httpbulb

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DynamicSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *DynamicSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *DynamicSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *DynamicSuite) TestStream() {

	type serverResponse struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}

	numMessages := 3
	apiURL := fmt.Sprintf("%s/stream/%d", s.testServer.URL, numMessages)

	req, err := http.NewRequest("GET", apiURL, nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	totalMsg := 0
	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < numMessages; i++ {
		rawMsg := scanner.Bytes()
		jsonMsg := &serverResponse{}
		json.Unmarshal(rawMsg, jsonMsg)
		if jsonMsg.URL != "" {
			totalMsg++
		}
	}

	assert.NoError(s.T(), scanner.Err())

	assert.Equal(s.T(), numMessages, totalMsg)

}

func (s *DynamicSuite) TestDelay() {
	d := 2

	apiURL := fmt.Sprintf("%s/delay/%d", s.testServer.URL, d)
	started := time.Now()

	req, err := http.NewRequest("GET", apiURL, nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	elapsed := time.Since(started)
	delay := time.Second * time.Duration(d)
	assert.GreaterOrEqual(s.T(), elapsed, delay)

}

func (s *DynamicSuite) TestBase64Decode() {

	encoded := base64.URLEncoding.EncodeToString([]byte("base64-decode test\n"))

	apiURL := fmt.Sprintf("%s/base64/%s", s.testServer.URL, encoded)

	req, err := http.NewRequest("GET", apiURL, nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), "base64-decode test\n", string(body))

}

func TestDynamicSuite(t *testing.T) {
	suite.Run(t, new(DynamicSuite))
}
