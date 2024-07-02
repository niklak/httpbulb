package httpbulb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
		name           string
		algorithm      string
		qop            string
		wantAuth       bool
		wantStatusCode int
		staleAfter     string
		skippingKeys   []string
	}

	tests := []testArgs{
		{name: "valid MD5 digest",
			algorithm: "MD5", qop: "auth", wantAuth: true, wantStatusCode: http.StatusOK},
		{name: "staled MD5 digest",
			algorithm: "MD5", qop: "auth", wantStatusCode: http.StatusUnauthorized, staleAfter: "0"},
		{name: "valid SHA-256 digest",
			algorithm: "SHA-256", qop: "auth", wantAuth: true, wantStatusCode: http.StatusOK},
		{name: "valid SHA-512 digest",
			algorithm: "SHA-512", qop: "auth", wantAuth: true, wantStatusCode: http.StatusOK},

		{name: "missing nc",
			algorithm: "MD5", qop: "auth", wantStatusCode: http.StatusUnauthorized, skippingKeys: []string{"nc"}},
		{name: "missing cnonce",
			algorithm: "MD5", qop: "auth", wantStatusCode: http.StatusUnauthorized, skippingKeys: []string{"cnonce"}},
	}

	username := "mememe"
	password := "mymymy"

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			uri := fmt.Sprintf("/digest-auth/%s/%s/%s/%s", tt.qop, username, password, tt.algorithm)
			addr := fmt.Sprintf("%s%s", s.testServer.URL, uri)
			credentials := map[string]string{
				"username":  username,
				"realm":     "httpbulb",
				"qop":       tt.qop,
				"uri":       uri,
				"nonce":     "dcd98b7102dd2f0e8b11d0f600bfb0c093",
				"nc":        "00000001",
				"cnonce":    "0a4f113b",
				"algorithm": tt.algorithm,
			}

			for _, key := range tt.skippingKeys {
				delete(credentials, key)
			}

			dig := (&digestCredentials{}).fromMap(credentials)

			digestResp := compileDigestResponse(dig, password, http.MethodGet, uri)

			credentials["response"] = digestResp

			credentialList := make([]string, 0, len(credentials)+1)
			for k, v := range credentials {
				credentialList = append(credentialList, fmt.Sprintf(`%s="%s"`, k, v))
			}

			auth := strings.Join(credentialList, ", ")

			req, err := http.NewRequest("GET", addr, nil)
			require.NoError(t, err)
			req.Header.Add("Authorization", fmt.Sprintf(`Digest %s`, auth))

			if tt.staleAfter != "" {
				req.AddCookie(&http.Cookie{
					Name:  "stale_after",
					Value: tt.staleAfter,
					Path:  "/",
				})
			}

			resp, err := s.client.Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)

			result := &serverResponse{}

			if len(body) > 0 {
				err = json.Unmarshal(body, result)
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantAuth, result.Authenticated)
		})

	}

}

func TestAuthDigestSuite(t *testing.T) {
	suite.Run(t, new(AuthDigestSuite))
}
