package httpbulb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		apiPath         string
		wantContentType string
		wantStatusCode  int
	}

	tests := []testArgs{
		{apiPath: "/image/svg", wantContentType: "image/svg+xml", wantStatusCode: http.StatusOK},
		{apiPath: "/image/png", wantContentType: "image/png", wantStatusCode: http.StatusOK},
		{apiPath: "/image/jpeg", wantContentType: "image/jpeg", wantStatusCode: http.StatusOK},
		{apiPath: "/image/webp", wantContentType: "image/webp", wantStatusCode: http.StatusOK},
		{apiPath: "/image/gif", wantContentType: "text/plain; charset=utf-8", wantStatusCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		req, err := http.NewRequest("GET", s.testServer.URL+tt.apiPath, nil)
		s.Require().NoError(err)

		resp, err := s.client.Do(req)
		s.Require().NoError(err)

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		s.Require().Equal(tt.wantStatusCode, resp.StatusCode)
		s.Require().Equal(tt.wantContentType, resp.Header.Get("Content-Type"))
	}

}

func (s *ImageSuite) TestImageAccept() {

	type testArgs struct {
		accept          string
		wantContentType string
		wantStatusCode  int
	}

	tests := []testArgs{
		{accept: "image/svg+xml", wantContentType: "image/svg+xml", wantStatusCode: http.StatusOK},
		{accept: "image/png", wantContentType: "image/png", wantStatusCode: http.StatusOK},
		{accept: "image/jpeg", wantContentType: "image/jpeg", wantStatusCode: http.StatusOK},
		{accept: "image/webp", wantContentType: "image/webp", wantStatusCode: http.StatusOK},
		{accept: "image/gif", wantContentType: "application/json", wantStatusCode: http.StatusNotAcceptable},
		{accept: "image/*", wantContentType: "image/png", wantStatusCode: http.StatusOK},
	}
	for _, tt := range tests {
		req, err := http.NewRequest("GET", s.testServer.URL+"/image", nil)
		s.Require().NoError(err)
		req.Header.Set("Accept", tt.accept)

		resp, err := s.client.Do(req)
		s.Require().NoError(err)

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		s.Require().Equal(tt.wantStatusCode, resp.StatusCode)
		s.Require().Equal(tt.wantContentType, resp.Header.Get("Content-Type"))
	}

}

func TestImageSuite(t *testing.T) {
	suite.Run(t, new(ImageSuite))
}
