package httpbulb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CorsSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *CorsSuite) SetupSuite() {

	handleFunc := NewRouter(Cors)
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *CorsSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *CorsSuite) TestCors() {
	type testArgs struct {
		name        string
		method      string
		headers     http.Header
		wantHeaders http.Header
	}

	tests := []testArgs{
		{
			name:        "Test Access-Control-Allow-Origin",
			method:      http.MethodGet,
			headers:     http.Header{"Origin": []string{"https://example.com"}},
			wantHeaders: http.Header{"Access-Control-Allow-Origin": []string{"https://example.com"}},
		},
		{
			name:        "Test Access-Control-Allow-Origin All",
			method:      http.MethodGet,
			headers:     http.Header{},
			wantHeaders: http.Header{"Access-Control-Allow-Origin": []string{"*"}},
		},
		{
			name:        "Test Options Access-Control-Allow-Methods",
			method:      http.MethodOptions,
			headers:     http.Header{},
			wantHeaders: http.Header{"Access-Control-Allow-Methods": []string{"GET, POST, PUT, DELETE, PATCH, OPTIONS"}},
		},
		{
			name:        "Test Access-Control-Request-Headers",
			method:      http.MethodOptions,
			headers:     http.Header{"Access-Control-Request-Headers": []string{"X-Test"}},
			wantHeaders: http.Header{"Access-Control-Allow-Headers": []string{"X-Test"}},
		},
	}

	apiURL := s.testServer.URL + "/get"
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, apiURL, nil)
			require.NoError(t, err)
			req.Header = tt.headers
			resp, err := s.client.Do(req)
			require.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			assert.Subset(t, resp.Header, tt.wantHeaders)
		})
	}

}

func TestCorsSuiteSuite(t *testing.T) {
	suite.Run(t, new(CorsSuite))
}
