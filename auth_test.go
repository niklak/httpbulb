package httpbulb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

func (s *AuthSuite) TestBasicAuthOk() {

	type serverResponse struct {
		Authenticated bool   `json:"authenticated"`
		User          string `json:"user"`
	}

	user := "mememe"
	passwd := "mymymy"

	addr := fmt.Sprintf("%s/basic-auth/%s/%s", s.testServer.URL, user, passwd)
	req, err := http.NewRequest("GET", addr, nil)
	assert.NoError(s.T(), err)

	req.SetBasicAuth(user, passwd)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	result := &serverResponse{}
	err = json.Unmarshal(body, result)
	assert.NoError(s.T(), err)

	expected := &serverResponse{Authenticated: true, User: user}

	assert.Equal(s.T(), expected, result)

}

func (s *AuthSuite) TestBasicAuthErr() {

	user := "mememe"
	passwd := "mymymy"

	addr := fmt.Sprintf("%s/basic-auth/%s/%s", s.testServer.URL, user, passwd)
	req, err := http.NewRequest("GET", addr, nil)
	assert.NoError(s.T(), err)

	req.SetBasicAuth(user, "wrongpasswd")

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)

}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
