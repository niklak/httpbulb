package httpbulb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *AuthSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *AuthSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *AuthSuite) TestBasicAuth() {

	type testArgs struct {
		username       string
		password       string
		authUsername   string
		authPassword   string
		wantAuth       bool
		wantStatusCode int
	}

	type serverResponse struct {
		Authenticated bool   `json:"authenticated"`
		User          string `json:"user"`
	}

	tests := []testArgs{
		{
			username:       "user1234",
			password:       "password1234",
			authUsername:   "user1234",
			authPassword:   "password1234",
			wantAuth:       true,
			wantStatusCode: http.StatusOK,
		},
		{
			username:       "user1234",
			password:       "password1234",
			authUsername:   "user1234",
			authPassword:   "password4321",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		apiURL := fmt.Sprintf("%s/basic-auth/%s/%s", s.testServer.URL, tt.username, tt.password)
		req, err := http.NewRequest("GET", apiURL, nil)
		s.Require().NoError(err)
		req.SetBasicAuth(tt.authUsername, tt.authPassword)

		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		resp.Body.Close()

		s.Require().Equal(tt.wantStatusCode, resp.StatusCode)

		result := &serverResponse{}
		err = json.Unmarshal(body, result)
		s.Require().NoError(err)

		s.Require().Equal(tt.wantAuth, result.Authenticated)

	}

}

func (s *AuthSuite) TestHiddenBasicAuthErr() {

	user := "user1234"
	passwd := "password1234"

	addr := fmt.Sprintf("%s/hidden-basic-auth/%s/%s", s.testServer.URL, user, passwd)
	req, err := http.NewRequest("GET", addr, nil)
	s.Require().NoError(err)

	req.SetBasicAuth(user, "wrong-password")

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	s.Require().Equal(http.StatusNotFound, resp.StatusCode)

}

func (s *AuthSuite) TestBearerAuth() {

	type testArgs struct {
		authPrefix     string
		token          string
		wantAuth       bool
		wantStatusCode int
	}

	type serverResponse struct {
		Authenticated bool   `json:"authenticated"`
		Token         string `json:"token"`
	}

	tests := []testArgs{
		{authPrefix: "Bearer ", token: "1234567890", wantAuth: true, wantStatusCode: http.StatusOK},
		{authPrefix: "Token ", token: "1234567890", wantAuth: false, wantStatusCode: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		addr := fmt.Sprintf("%s/bearer", s.testServer.URL)
		req, err := http.NewRequest("GET", addr, nil)
		s.Require().NoError(err)

		req.Header.Set("Authorization", tt.authPrefix+tt.token)

		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		resp.Body.Close()

		s.Require().Equal(tt.wantStatusCode, resp.StatusCode)

		result := &serverResponse{}
		err = json.Unmarshal(body, result)
		s.Require().NoError(err)

		s.Require().Equal(tt.wantAuth, result.Authenticated)
	}

}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
