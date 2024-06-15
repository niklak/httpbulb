package httpbulb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
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

func (s *ResponseInspectionSuite) TestCache() {

	type testArgs struct {
		ifModifiedSince string
		ifNoneMatch     string
		wantStatusCode  int
	}

	tests := []testArgs{
		{
			wantStatusCode: http.StatusOK,
		},
		{
			ifModifiedSince: time.Now().Format(time.RFC1123),
			wantStatusCode:  http.StatusNotModified,
		},
		{
			ifNoneMatch:    uuid.New().String(),
			wantStatusCode: http.StatusNotModified,
		},
		{
			ifModifiedSince: time.Now().Format(time.RFC1123),
			ifNoneMatch:     uuid.New().String(),
			wantStatusCode:  http.StatusNotModified,
		},
	}
	for _, tt := range tests {
		apiURL := s.testServer.URL + "/cache"

		req, err := http.NewRequest("GET", apiURL, nil)
		assert.NoError(s.T(), err)

		if tt.ifModifiedSince != "" {
			req.Header.Set("If-Modified-Since", tt.ifModifiedSince)
		}
		if tt.ifNoneMatch != "" {
			req.Header.Set("If-None-Match", tt.ifNoneMatch)
		}

		resp, err := s.client.Do(req)
		assert.NoError(s.T(), err)

		io.Copy(io.Discard, resp.Body)

		resp.Body.Close()

		assert.Equal(s.T(), tt.wantStatusCode, resp.StatusCode)
	}

}

func (s *ResponseInspectionSuite) TestCacheControl() {

	type serverResponse struct {
		URL string `json:"url"`
	}

	value := "3600"
	apiURL := s.testServer.URL + "/cache/" + value

	req, err := http.NewRequest("GET", apiURL, nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), apiURL, result.URL)

	expectedCacheControl := fmt.Sprintf("public, max-age=%s", value)

	assert.Equal(s.T(), expectedCacheControl, resp.Header.Get("Cache-Control"))

}

func (s *ResponseInspectionSuite) TestEtag() {

	type testArgs struct {
		ifNoneMatch string
		ifMatch     string
		wantStatus  int
	}
	etag := uuid.New().String()

	tests := []testArgs{
		{
			wantStatus: http.StatusOK,
		},
		{
			ifNoneMatch: etag,
			wantStatus:  http.StatusNotModified,
		},
		{
			ifNoneMatch: "*",
			wantStatus:  http.StatusNotModified,
		},
		{
			ifNoneMatch: uuid.NewString(),
			wantStatus:  http.StatusOK,
		},
		{
			ifMatch:    etag,
			wantStatus: http.StatusOK,
		},
		{
			ifMatch:    "*",
			wantStatus: http.StatusOK,
		},
		{
			ifMatch:    uuid.NewString(),
			wantStatus: http.StatusPreconditionFailed,
		},
	}
	for _, tt := range tests {

		apiURL := s.testServer.URL + "/etag/" + etag

		req, err := http.NewRequest("GET", apiURL, nil)
		assert.NoError(s.T(), err)

		if tt.ifNoneMatch != "" {
			req.Header.Set("If-None-Match", tt.ifNoneMatch)
		}
		if tt.ifMatch != "" {
			req.Header.Set("If-Match", tt.ifMatch)
		}

		resp, err := s.client.Do(req)
		assert.NoError(s.T(), err)

		io.Copy(io.Discard, resp.Body)
		assert.NoError(s.T(), err)

		resp.Body.Close()

		assert.Equal(s.T(), tt.wantStatus, resp.StatusCode)
	}

}

func TestResponseInspectionSuite(t *testing.T) {
	suite.Run(t, new(ResponseInspectionSuite))
}
