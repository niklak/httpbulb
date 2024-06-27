package httpbulb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ImageSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *ImageSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *ImageSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *ImageSuite) TestImage() {

	type testArgs struct {
		name            string
		apiPath         string
		wantContentType string
		wantStatusCode  int
	}

	tests := []testArgs{
		{name: "svg", apiPath: "/image/svg", wantContentType: "image/svg+xml", wantStatusCode: http.StatusOK},
		{name: "png", apiPath: "/image/png", wantContentType: "image/png", wantStatusCode: http.StatusOK},
		{name: "jpeg", apiPath: "/image/jpeg", wantContentType: "image/jpeg", wantStatusCode: http.StatusOK},
		{name: "webp", apiPath: "/image/webp", wantContentType: "image/webp", wantStatusCode: http.StatusOK},
		{name: "not found", apiPath: "/image/gif", wantContentType: "text/plain; charset=utf-8", wantStatusCode: http.StatusNotFound},
		{name: "avif", apiPath: "/image/avif", wantContentType: "image/avif", wantStatusCode: http.StatusOK},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", s.testServer.URL+tt.apiPath, nil)
			require.NoError(t, err)

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)
			require.Equal(t, tt.wantContentType, resp.Header.Get("Content-Type"))
		})

	}

}

func (s *ImageSuite) TestImageAccept() {

	type testArgs struct {
		name            string
		accept          string
		wantContentType string
		wantStatusCode  int
	}

	tests := []testArgs{
		{name: "svg", accept: "image/svg+xml", wantContentType: "image/svg+xml", wantStatusCode: http.StatusOK},
		{name: "png", accept: "image/png", wantContentType: "image/png", wantStatusCode: http.StatusOK},
		{name: "jpeg", accept: "image/jpeg", wantContentType: "image/jpeg", wantStatusCode: http.StatusOK},
		{name: "webp", accept: "image/webp", wantContentType: "image/webp", wantStatusCode: http.StatusOK},
		{name: "gif", accept: "image/gif", wantContentType: "application/json", wantStatusCode: http.StatusNotAcceptable},
		{name: "gif", accept: "text/plain", wantContentType: "application/json", wantStatusCode: http.StatusNotAcceptable},
		{name: "any", accept: "image/*", wantContentType: "image/png", wantStatusCode: http.StatusOK},
		{name: "avif", accept: "image/avif, image/webp, */*;q=0.8", wantContentType: "image/avif", wantStatusCode: http.StatusOK},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", s.testServer.URL+"/image", nil)
			require.NoError(t, err)
			req.Header.Set("Accept", tt.accept)

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)
			require.Equal(t, tt.wantContentType, resp.Header.Get("Content-Type"))
		})

	}

}

func TestImageSuite(t *testing.T) {
	suite.Run(t, new(ImageSuite))
}
