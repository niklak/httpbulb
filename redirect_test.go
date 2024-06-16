package httpbulb

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RedirectSuite struct {
	suite.Suite
	testServer        *httptest.Server
	clientNoRedirect  *http.Client
	clientOneRedirect *http.Client
}

func (s *RedirectSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.clientOneRedirect = &http.Client{}
	s.clientOneRedirect.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 1 {
			return http.ErrUseLastResponse
		}
		return nil
	}

	s.clientNoRedirect = &http.Client{}

	// client will not follow redirects
	s.clientNoRedirect.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (s *RedirectSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *RedirectSuite) TestRedirectTo() {
	var methods = []string{"GET", "DELETE", "POST", "PUT", "PATCH"}
	for _, method := range methods {
		headers := http.Header{}
		s.T().Run(method, func(t *testing.T) {
			dstURL := s.testServer.URL + "/" + strings.ToLower(method)

			apiURL, err := url.Parse(s.testServer.URL)
			require.NoError(t, err)
			apiURL.Path = "/redirect-to"
			var body io.Reader
			switch method {
			case http.MethodGet, http.MethodDelete:
				query := url.Values{}
				query.Set("url", dstURL)
				query.Set("status", strconv.Itoa(http.StatusMovedPermanently))
				apiURL.RawQuery = query.Encode()
			case http.MethodPost, http.MethodPut, http.MethodPatch:
				form := url.Values{}
				form.Set("url", dstURL)
				form.Set("status", strconv.Itoa(http.StatusMovedPermanently))
				body = strings.NewReader(form.Encode())
				headers.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			req, err := http.NewRequest(method, apiURL.String(), body)
			require.NoError(t, err)
			req.Header = headers
			resp, err := s.clientNoRedirect.Do(req)
			require.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			require.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
			require.Equal(t, dstURL, resp.Header.Get("Location"))
		})
	}

}

func (s *RedirectSuite) TestRedirects() {

	type testArgs struct {
		apiURL         string
		wantLocation   string
		wantStatusCode int
	}

	tests := []testArgs{
		{
			apiURL:         fmt.Sprintf("%s/redirect/3", s.testServer.URL),
			wantLocation:   "/relative-redirect/2",
			wantStatusCode: http.StatusFound,
		},
		{
			apiURL:         fmt.Sprintf("%s/redirect/0", s.testServer.URL),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			apiURL:         fmt.Sprintf("%s/redirect/a", s.testServer.URL),
			wantStatusCode: http.StatusNotFound,
		},
		{
			apiURL:         fmt.Sprintf("%s/redirect/3?absolute=true", s.testServer.URL),
			wantLocation:   fmt.Sprintf("%s/absolute-redirect/2", s.testServer.URL),
			wantStatusCode: http.StatusFound,
		},
		{
			apiURL:         fmt.Sprintf("%s/relative-redirect/3", s.testServer.URL),
			wantLocation:   "/relative-redirect/2",
			wantStatusCode: http.StatusFound,
		},
		{
			apiURL:         fmt.Sprintf("%s/relative-redirect/1", s.testServer.URL),
			wantLocation:   "/get",
			wantStatusCode: http.StatusFound,
		},
		{
			apiURL:         fmt.Sprintf("%s/absolute-redirect/3", s.testServer.URL),
			wantLocation:   fmt.Sprintf("%s/absolute-redirect/2", s.testServer.URL),
			wantStatusCode: http.StatusFound,
		},
		{
			apiURL:         fmt.Sprintf("%s/absolute-redirect/1", s.testServer.URL),
			wantLocation:   fmt.Sprintf("%s/get", s.testServer.URL),
			wantStatusCode: http.StatusFound,
		},
	}

	for _, tt := range tests {
		req, err := http.NewRequest(http.MethodGet, tt.apiURL, nil)
		s.Require().NoError(err)
		resp, err := s.clientOneRedirect.Do(req)
		s.Require().NoError(err)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		s.Require().Equal(tt.wantStatusCode, resp.StatusCode)
		// as this test follows only for one redirect, the location should be: /relative-redirect/2
		s.Require().Equal(tt.wantLocation, resp.Header.Get("Location"))
	}
}

func TestRedirectSuite(t *testing.T) {
	suite.Run(t, new(RedirectSuite))
}

func Test_redirectHandle(t *testing.T) {
	type args struct {
		w        http.ResponseWriter
		r        *http.Request
		absolute bool
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redirectHandle(tt.args.w, tt.args.r, tt.args.absolute)
		})
	}
}
