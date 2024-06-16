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
	"github.com/stretchr/testify/require"
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
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

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

	s.Require().NoError(scanner.Err())

	s.Require().Equal(numMessages, totalMsg)

}

func (s *DynamicSuite) TestDelay() {
	d := 2

	apiURL := fmt.Sprintf("%s/delay/%d", s.testServer.URL, d)
	started := time.Now()

	req, err := http.NewRequest("GET", apiURL, nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	elapsed := time.Since(started)
	delay := time.Second * time.Duration(d)
	s.Require().GreaterOrEqual(elapsed, delay)

}

func (s *DynamicSuite) TestBase64Decode() {

	encoded := base64.URLEncoding.EncodeToString([]byte("base64-decode test\n"))

	apiURL := fmt.Sprintf("%s/base64/%s", s.testServer.URL, encoded)

	req, err := http.NewRequest("GET", apiURL, nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	s.Require().Equal("base64-decode test\n", string(body))

}

func (s *DynamicSuite) TestRandomBytes() {

	type testArgs struct {
		name     string
		byteSize int
	}

	tests := []testArgs{
		{"bytes 0", 0},
		{"bytes 10", 10},
		{"bytes 512", 512},
	}

	for _, tt := range tests {

		s.T().Run(tt.name, func(t *testing.T) {

			apiURL := fmt.Sprintf("%s/bytes/%d", s.testServer.URL, tt.byteSize)

			req, err := http.NewRequest("GET", apiURL, nil)
			require.NoError(t, err)

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, tt.byteSize, len(body))
		})

	}

}

func (s *DynamicSuite) TestStreamRandomBytes() {

	type testArgs struct {
		name      string
		chunkSize int
	}

	numBytes := 1024

	// chunk sizes:
	// 0 will be used 1 by default -- chunk by 1 byte
	// 24 bytes as chunk, the last chunk will be lesser then others
	// 512 bytes -- nothing special
	// 2028 bytes -- this chunk size bigger then the total bytes, so the body response will be sent in one chunk

	tests := []testArgs{
		{"chunk size 0", 0},
		{"chunk size 24", 24},
		{"chunk size 512", 512},
		{"chunk size 2028", 2028},
	}

	for _, tt := range tests {

		s.T().Run(tt.name, func(t *testing.T) {
			apiURL := fmt.Sprintf("%s/stream-bytes/%d?chunk_size=%d", s.testServer.URL, numBytes, tt.chunkSize)

			req, err := http.NewRequest("GET", apiURL, nil)
			s.Require().NoError(err)

			resp, err := s.client.Do(req)
			s.Require().NoError(err)

			s.Require().Equal(http.StatusOK, resp.StatusCode)

			defer resp.Body.Close()

			s.Require().NoError(err)

			body, err := io.ReadAll(resp.Body)

			s.Require().NoError(err)
			receivedBytes := len(body)

			s.Require().Equal(numBytes, receivedBytes)
		})

	}
}

func (s *DynamicSuite) TestUUID() {

	apiURL := fmt.Sprintf("%s/uuid", s.testServer.URL)

	req, err := http.NewRequest("GET", apiURL, nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	var uuidResp UUIDResponse
	err = json.NewDecoder(resp.Body).Decode(&uuidResp)
	s.Require().NoError(err)

	_, err = uuid.Parse(uuidResp.UUID)
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	started := time.Now()

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	elapsed := time.Since(started)
	expected := time.Second * (time.Duration(delay) + time.Duration(dur))

	s.Require().LessOrEqual(expected, elapsed)

	s.Require().Equal(strings.Repeat("*", numBytes), string(body))

}

func (s *DynamicSuite) TestLinkPage() {

	numLinks := 3
	offset := 1

	apiURL := fmt.Sprintf("%s/links/%d/%d", s.testServer.URL, numLinks, offset)

	req, err := http.NewRequest("GET", apiURL, nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	expected := `<html><head><title>Links</title></head><body><a href='/links/3/0'>0</a> 1 <a href='/links/3/2'>2</a> </body></html>`

	s.Require().Equal(expected, string(body))
}

func (s *DynamicSuite) TestLinks() {

	numLinks := 3

	apiURL := fmt.Sprintf("%s/links/%d", s.testServer.URL, numLinks)

	req, err := http.NewRequest("GET", apiURL, nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	expected := `<html><head><title>Links</title></head><body>0 <a href='/links/3/1'>1</a> <a href='/links/3/2'>2</a> </body></html>`

	s.Require().Equal(expected, string(body))
}

func (s *DynamicSuite) TestRange() {

	numBytes := 30
	rangeHeader := "bytes=10-20"
	//duration is used only to calculate a pause per byte
	dur := 2
	apiURL := fmt.Sprintf("%s/range/%d?duration=%d", s.testServer.URL, numBytes, dur)

	req, err := http.NewRequest("GET", apiURL, nil)
	s.Require().NoError(err)

	req.Header.Set("Range", rangeHeader)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	s.Require().Equal(http.StatusPartialContent, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	expectedHeaders := http.Header{
		"Etag":           []string{fmt.Sprintf("range%d", numBytes)},
		"Content-Length": []string{"11"},
		"Content-Range":  []string{"bytes 10-20/30"},
	}

	s.Require().Equal("klmnopqrstu", string(body))
	s.Require().Subset(resp.Header, expectedHeaders)

}

func TestDynamicSuite(t *testing.T) {
	suite.Run(t, new(DynamicSuite))
}
