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

	type testArgs struct {
		name            string
		method          string
		statusCode      int
		wantStatusCode  int
		noCheckLocation bool
		ct              string
	}

	tests := []testArgs{
		{name: "GET", method: http.MethodGet, statusCode: 301, wantStatusCode: 301},
		{name: "DELETE", method: http.MethodDelete, statusCode: 301, wantStatusCode: 301},
		{name: "POST", method: http.MethodPost, statusCode: 307, wantStatusCode: 307},
		{name: "PUT", method: http.MethodPut, statusCode: 307, wantStatusCode: 307},
		{name: "PATCH", method: http.MethodPatch, statusCode: 307, wantStatusCode: 307},
		{name: "GET ignore 200", method: http.MethodGet, statusCode: 200, wantStatusCode: 302},
		{
			name:            "POST bad content-type",
			method:          http.MethodPost,
			wantStatusCode:  400,
			ct:              "x-www-form-urlencoded",
			noCheckLocation: true,
		},
	}

	for _, tt := range tests {

		headers := http.Header{}
		s.T().Run(tt.name, func(t *testing.T) {
			dstURL := s.testServer.URL + "/" + strings.ToLower(tt.method)

			apiURL, err := url.Parse(s.testServer.URL)
			require.NoError(t, err)
			apiURL.Path = "/redirect-to"
			var body io.Reader
			switch tt.method {
			case http.MethodGet, http.MethodDelete:
				query := url.Values{}
				query.Set("url", dstURL)
				query.Set("status", strconv.Itoa(tt.wantStatusCode))
				apiURL.RawQuery = query.Encode()
			case http.MethodPost, http.MethodPut, http.MethodPatch:
				form := url.Values{}
				form.Set("url", dstURL)
				form.Set("status", strconv.Itoa(tt.wantStatusCode))
				body = strings.NewReader(form.Encode())
				ct := tt.ct
				if ct == "" {
					ct = "application/x-www-form-urlencoded"
				}
				headers.Set("Content-Type", ct)
			}
			req, err := http.NewRequest(tt.method, apiURL.String(), body)
			require.NoError(t, err)
			req.Header = headers
			resp, err := s.clientNoRedirect.Do(req)
			require.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			require.Equal(t, tt.wantStatusCode, resp.StatusCode)
			if tt.noCheckLocation {
				return
			}
			require.Equal(t, dstURL, resp.Header.Get("Location"))
		})
	}

}

func (s *RedirectSuite) TestRedirects() {

	type testArgs struct {
		name           string
		apiURL         string
		wantLocation   string
		wantStatusCode int
	}

	tests := []testArgs{
		{
			name:           "redirect to relative",
			apiURL:         fmt.Sprintf("%s/redirect/3", s.testServer.URL),
			wantLocation:   "/relative-redirect/2",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "bad redirect",
			apiURL:         fmt.Sprintf("%s/redirect/0", s.testServer.URL),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "not found",
			apiURL:         fmt.Sprintf("%s/redirect/a", s.testServer.URL),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "redirect to absolute",
			apiURL:         fmt.Sprintf("%s/redirect/3?absolute=true", s.testServer.URL),
			wantLocation:   fmt.Sprintf("%s/absolute-redirect/2", s.testServer.URL),
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "relative redirect",
			apiURL:         fmt.Sprintf("%s/relative-redirect/3", s.testServer.URL),
			wantLocation:   "/relative-redirect/2",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "successful relative redirect",
			apiURL:         fmt.Sprintf("%s/relative-redirect/1", s.testServer.URL),
			wantLocation:   "/get",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "absolute redirect",
			apiURL:         fmt.Sprintf("%s/absolute-redirect/3", s.testServer.URL),
			wantLocation:   fmt.Sprintf("%s/absolute-redirect/2", s.testServer.URL),
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "successful absolute redirect",
			apiURL:         fmt.Sprintf("%s/absolute-redirect/1", s.testServer.URL),
			wantLocation:   fmt.Sprintf("%s/get", s.testServer.URL),
			wantStatusCode: http.StatusFound,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.apiURL, nil)
			require.NoError(t, err)
			resp, err := s.clientOneRedirect.Do(req)
			require.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			require.Equal(t, tt.wantStatusCode, resp.StatusCode)
			// as this test follows only for one redirect, the location should be: /relative-redirect/n-1
			require.Equal(t, tt.wantLocation, resp.Header.Get("Location"))
		})

	}
}

func TestRedirectSuite(t *testing.T) {
	suite.Run(t, new(RedirectSuite))
}
