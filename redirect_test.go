package httpbulb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedirectSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *RedirectSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient

	// client will not follow redirects
	s.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (s *RedirectSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *RedirectSuite) TestRedirectParam() {
	var methods = []string{"GET", "DELETE", "POST", "PUT", "PATCH"}
	for _, method := range methods {
		headers := http.Header{}
		s.T().Run(method, func(t *testing.T) {
			dstURL := s.testServer.URL + "/" + strings.ToLower(method)

			apiURL, err := url.Parse(s.testServer.URL)
			assert.NoError(t, err)
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
			assert.NoError(t, err)
			req.Header = headers
			resp, err := s.client.Do(req)
			assert.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
			assert.Equal(t, dstURL, resp.Header.Get("Location"))
		})
	}

}

func TestRedirectSuite(t *testing.T) {
	suite.Run(t, new(RedirectSuite))
}
