package httpbulb

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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

func (s *DynamicSuite) TestRandomBytes() {

	bytesSizes := []int{0, 10, 512}

	for _, numBytes := range bytesSizes {
		apiURL := fmt.Sprintf("%s/bytes/%d", s.testServer.URL, numBytes)

		req, err := http.NewRequest("GET", apiURL, nil)
		assert.NoError(s.T(), err)

		resp, err := s.client.Do(req)
		assert.NoError(s.T(), err)

		assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		assert.NoError(s.T(), err)

		assert.Equal(s.T(), numBytes, len(body))
	}

}

func (s *DynamicSuite) TestStreamRandomBytes() {

	numBytes := 1024

	// chunk sizes:
	// 0 will be used 1 by default -- chunk by 1 byte
	// 24 bytes as chunk, the last chunk will be lesser then others
	// 512 bytes -- nothing special
	// 2028 bytes -- this chunk size bigger then the total bytes, so the body response will be sent in one chunk
	chunkSizes := []int{0, 24, 512, 2028}

	for _, chunkSize := range chunkSizes {
		apiURL := fmt.Sprintf("%s/stream-bytes/%d?chunk_size=%d", s.testServer.URL, numBytes, chunkSize)

		req, err := http.NewRequest("GET", apiURL, nil)
		assert.NoError(s.T(), err)

		resp, err := s.client.Do(req)
		assert.NoError(s.T(), err)

		assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		assert.NoError(s.T(), err)

		body, err := io.ReadAll(resp.Body)

		assert.NoError(s.T(), err)
		receivedBytes := len(body)

		assert.Equal(s.T(), numBytes, receivedBytes)
	}
}

func (s *DynamicSuite) TestUUID() {

	apiURL := fmt.Sprintf("%s/uuid", s.testServer.URL)

	req, err := http.NewRequest("GET", apiURL, nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	var uuidResp UUIDResponse
	err = json.NewDecoder(resp.Body).Decode(&uuidResp)
	assert.NoError(s.T(), err)

	_, err = uuid.Parse(uuidResp.UUID)
	assert.NoError(s.T(), err)

}

func (s *DynamicSuite) TestDrip() {

	delay := 2
	dur := 2
	numBytes := 10
	apiURL := fmt.Sprintf(
		"%s/drip?numbytes=%d&delay=%d&duration=%d",
		s.testServer.URL, numBytes, delay, dur,
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	assert.NoError(s.T(), err)

	started := time.Now()

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	elapsed := time.Since(started)
	expected := time.Second * (time.Duration(delay) + time.Duration(dur))

	assert.LessOrEqual(s.T(), expected, elapsed)

	assert.Equal(s.T(), strings.Repeat("*", numBytes), string(body))

}

func TestDynamicSuite(t *testing.T) {
	suite.Run(t, new(DynamicSuite))
}
