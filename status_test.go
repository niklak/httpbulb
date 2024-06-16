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

type StatusCodeSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *StatusCodeSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *StatusCodeSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *StatusCodeSuite) TestNotACode() {
	testURL := fmt.Sprintf("%s/status/bad", s.testServer.URL)

	req, err := http.NewRequest("GET", testURL, nil)
	s.Require().NoError(err)
	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *StatusCodeSuite) TestStatusCodes() {

	type testArgs struct {
		method         string
		encodePath     bool
		statusCodes    []int
		wantStatusCode int
	}

	tests := []testArgs{
		{method: "GET", statusCodes: []int{100}, wantStatusCode: 400},
		{method: "GET", statusCodes: []int{200, 403, 500}, encodePath: true},
		{method: "DELETE", statusCodes: []int{200, 403, 500}},
		{method: "PATCH", statusCodes: []int{200, 403, 500}},
		{method: "POST", statusCodes: []int{200, 403, 500}},
		{method: "PUT", statusCodes: []int{200, 403, 500}},
		{method: "GET", statusCodes: []int{444}},
		{method: "GET", statusCodes: []int{600}, wantStatusCode: 400},
		{method: "Get", statusCodes: []int{200, 403, 500}, wantStatusCode: 405},
	}
	for _, tt := range tests {
		var codes []string

		for _, code := range tt.statusCodes {
			codes = append(codes, strconv.Itoa(code))
		}
		codesPath := strings.Join(codes, ",")

		if tt.encodePath {
			codesPath = url.QueryEscape(codesPath)
		}

		testURL := fmt.Sprintf("%s/status/%s", s.testServer.URL, codesPath)

		req, err := http.NewRequest(tt.method, testURL, nil)
		s.Require().NoError(err)
		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if tt.wantStatusCode != 0 {
			s.Require().Equal(tt.wantStatusCode, resp.StatusCode)
			return
		}
		require.Contains(s.T(), tt.statusCodes, resp.StatusCode)

	}

}

func TestStatusCodeSuite(t *testing.T) {
	suite.Run(t, new(StatusCodeSuite))
}
