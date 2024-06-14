package httpbulb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthDigestSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *AuthDigestSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	s.client = http.DefaultClient
}

func (s *AuthDigestSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *AuthDigestSuite) TestDigestAuth() {
	type serverResponse struct {
		Authenticated bool   `json:"authenticated"`
		User          string `json:"user"`
	}

	type testArgs struct {
		algorithm      string
		qop            string
		wantAuth       bool
		wantStatusCode int
		staleAfter     string
	}

	tests := []testArgs{
		{algorithm: "MD5", qop: "auth", wantAuth: true, wantStatusCode: http.StatusOK},
		{algorithm: "MD5", qop: "auth", wantAuth: false, wantStatusCode: http.StatusUnauthorized, staleAfter: "0"},
		{algorithm: "SHA-256", qop: "auth", wantAuth: true, wantStatusCode: http.StatusOK},
		{algorithm: "SHA-512", qop: "auth", wantAuth: true, wantStatusCode: http.StatusOK},
	}

	username := "mememe"
	password := "mymymy"

	for _, tt := range tests {
		addr := fmt.Sprintf(
			"%s/digest-auth/%s/%s/%s/%s",
			s.testServer.URL, tt.qop, username, password, tt.algorithm)
		credentials := map[string]string{
			"username":  username,
			"realm":     "httpbulb",
			"qop":       tt.qop,
			"uri":       addr,
			"nonce":     "dcd98b7102dd2f0e8b11d0f600bfb0c093",
			"nc":        "00000001",
			"cnonce":    "0a4f113b",
			"algorithm": tt.algorithm,
		}

		digestResp := compileDigestResponse(credentials, password, http.MethodGet, addr)

		credentials["response"] = digestResp

		credentialList := make([]string, 0, len(credentials)+1)
		for k, v := range credentials {
			credentialList = append(credentialList, fmt.Sprintf(`%s="%s"`, k, v))
		}

		auth := strings.Join(credentialList, ", ")

		req, err := http.NewRequest("GET", addr, nil)
		assert.NoError(s.T(), err)

		req.Header.Add("Authorization", fmt.Sprintf(`Digest %s`, auth))

		if tt.staleAfter != "" {
			req.AddCookie(&http.Cookie{
				Name:  "stale_after",
				Value: tt.staleAfter,
				Path:  "/",
			})
		}

		resp, err := s.client.Do(req)
		assert.NoError(s.T(), err)

		assert.Equal(s.T(), tt.wantStatusCode, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(s.T(), err)
		resp.Body.Close()

		result := &serverResponse{}

		if len(body) > 0 {
			err = json.Unmarshal(body, result)
			assert.NoError(s.T(), err)

		}

		assert.Equal(s.T(), tt.wantAuth, result.Authenticated)
	}

}

func TestAuthDigestSuite(t *testing.T) {
	suite.Run(t, new(AuthDigestSuite))
}
