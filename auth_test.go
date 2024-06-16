package httpbulb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
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
		name           string
		apiPath        string
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
			name:           "valid basic auth",
			apiPath:        "basic-auth",
			username:       "user1234",
			password:       "password1234",
			authUsername:   "user1234",
			authPassword:   "password1234",
			wantAuth:       true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid basic auth",
			apiPath:        "basic-auth",
			username:       "user1234",
			password:       "password1234",
			authUsername:   "user1234",
			authPassword:   "password4321",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid hidden basic auth",
			apiPath:        "hidden-basic-auth",
			username:       "user1234",
			password:       "password1234",
			authUsername:   "user1234",
			authPassword:   "password4321",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {

		s.T().Run(tt.name, func(t *testing.T) {
			apiURL := fmt.Sprintf("%s/%s/%s/%s", s.testServer.URL, tt.apiPath, tt.username, tt.password)
			req, err := http.NewRequest("GET", apiURL, nil)
			require.NoError(t, err)
			req.SetBasicAuth(tt.authUsername, tt.authPassword)

			resp, err := s.client.Do(req)
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)

			result := &serverResponse{}
			err = json.Unmarshal(body, result)
			require.NoError(t, err)
			require.Equal(t, tt.wantAuth, result.Authenticated)
		})

	}

}

func (s *AuthSuite) TestBearerAuth() {

	type testArgs struct {
		name           string
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
		{name: "valid bearer", authPrefix: "Bearer ", token: "1234567890", wantAuth: true, wantStatusCode: http.StatusOK},
		{name: "invalid bearer", authPrefix: "Token ", token: "1234567890", wantAuth: false, wantStatusCode: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			addr := fmt.Sprintf("%s/bearer", s.testServer.URL)
			req, err := http.NewRequest("GET", addr, nil)
			require.NoError(t, err)

			req.Header.Set("Authorization", tt.authPrefix+tt.token)

			resp, err := s.client.Do(req)
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)

			result := &serverResponse{}
			err = json.Unmarshal(body, result)
			require.NoError(t, err)

			require.Equal(t, tt.wantAuth, result.Authenticated)
		})

	}

}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
