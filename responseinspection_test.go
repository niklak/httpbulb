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
	"github.com/stretchr/testify/require"
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
	s.Require().NoError(err)
	apiURL.Path = "/response-headers"

	query := url.Values{}
	query.Add("x-test-header", "1")
	query.Add("X-Test-Header", "2")
	query.Add("X-Test-Header", "3")
	apiURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	resp.Body.Close()

	result := http.Header{}

	err = json.Unmarshal(body, &result)
	s.Require().NoError(err)

	expectedHeaderValue := []string{"1", "2", "3"}
	assert.Subset(s.T(), expectedHeaderValue, result["X-Test-Header"])
	assert.Len(s.T(), result["X-Test-Header"], len(expectedHeaderValue))

}

func (s *ResponseInspectionSuite) TestCache() {

	type testArgs struct {
		name            string
		ifModifiedSince string
		ifNoneMatch     string
		wantStatusCode  int
	}

	tests := []testArgs{
		{
			name:           "no cache 200",
			wantStatusCode: http.StatusOK,
		},
		{
			name:            "If-Modified-Since 304",
			ifModifiedSince: time.Now().Format(time.RFC1123),
			wantStatusCode:  http.StatusNotModified,
		},
		{
			name:           "If-None-Match 304",
			ifNoneMatch:    uuid.New().String(),
			wantStatusCode: http.StatusNotModified,
		},
		{
			name:            "If-Modified-Since and If-None-Match 304",
			ifModifiedSince: time.Now().Format(time.RFC1123),
			ifNoneMatch:     uuid.New().String(),
			wantStatusCode:  http.StatusNotModified,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			apiURL := s.testServer.URL + "/cache"

			req, err := http.NewRequest("GET", apiURL, nil)
			require.NoError(t, err)

			if tt.ifModifiedSince != "" {
				req.Header.Set("If-Modified-Since", tt.ifModifiedSince)
			}
			if tt.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatch)
			}

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			io.Copy(io.Discard, resp.Body)

			resp.Body.Close()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})

	}

}

func (s *ResponseInspectionSuite) TestCacheControl() {

	type serverResponse struct {
		URL string `json:"url"`
	}

	value := "3600"
	apiURL := s.testServer.URL + "/cache/" + value

	req, err := http.NewRequest("GET", apiURL, nil)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	result := &serverResponse{}

	err = json.Unmarshal(body, result)
	s.Require().NoError(err)

	s.Require().Equal(apiURL, result.URL)

	expectedCacheControl := fmt.Sprintf("public, max-age=%s", value)

	s.Require().Equal(expectedCacheControl, resp.Header.Get("Cache-Control"))

}

func (s *ResponseInspectionSuite) TestEtag() {

	type testArgs struct {
		name        string
		ifNoneMatch string
		ifMatch     string
		wantStatus  int
	}
	etag := uuid.New().String()

	tests := []testArgs{
		{
			name:       "200",
			wantStatus: http.StatusOK,
		},
		{
			name:        "If-None-Match valid etag",
			ifNoneMatch: etag,
			wantStatus:  http.StatusNotModified,
		},
		{
			name:        "If-None-Match *",
			ifNoneMatch: "*",
			wantStatus:  http.StatusNotModified,
		},
		{
			name:        "If-None-Match invalid etag",
			ifNoneMatch: uuid.NewString(),
			wantStatus:  http.StatusOK,
		},
		{
			name:       "If-Match valid etag",
			ifMatch:    etag,
			wantStatus: http.StatusOK,
		},
		{
			name:       "If-Match * 200",
			ifMatch:    "*",
			wantStatus: http.StatusOK,
		},
		{
			name:       "If-Match invalid etag",
			ifMatch:    uuid.NewString(),
			wantStatus: http.StatusPreconditionFailed,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			apiURL := s.testServer.URL + "/etag/" + etag

			req, err := http.NewRequest("GET", apiURL, nil)
			require.NoError(t, err)

			if tt.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatch)
			}
			if tt.ifMatch != "" {
				req.Header.Set("If-Match", tt.ifMatch)
			}

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			io.Copy(io.Discard, resp.Body)
			require.NoError(t, err)

			resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)
		})

	}

}

func TestResponseInspectionSuite(t *testing.T) {
	suite.Run(t, new(ResponseInspectionSuite))
}
