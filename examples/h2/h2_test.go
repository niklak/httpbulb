package h2

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/niklak/httpbulb"
	"github.com/stretchr/testify/require"
)

func Test_Http2Client(t *testing.T) {
	// This test checks http client configuration to send HTTP/2.0 requests to the server with tls
	type testArgs struct {
		name          string
		client        *http.Client
		wantProto     string
		wantClientErr bool
	}

	// a representation of a response from the testing server,
	// for the test i require only an `url` and a `proto` fields, so i skip the rest.
	type serverResponse struct {
		URL   string `json:"url"`
		Proto string `json:"proto"`
	}

	// starting a tls test server with HTTP/2.0 support
	handleFunc := httpbulb.NewRouter()
	testServer := httptest.NewUnstartedServer(handleFunc)
	testServer.EnableHTTP2 = true
	testServer.StartTLS()
	defer testServer.Close()

	// prepare the url
	u, err := url.Parse(testServer.URL)
	require.NoError(t, err)
	u.Path = "/get"
	testURL := u.String()

	// this client is forced to use HTTP/2.0
	h2Client := testServer.Client()

	// this client is expected to fail, because i didn't install CA certificate for this server
	h1Client := &http.Client{}

	// a client that skips TLS
	h1ClientInsecure := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	tests := []testArgs{
		{name: "HTTP/2.0", client: h2Client, wantProto: "HTTP/2.0"},
		{name: "HTTP/1.1 - Insecure", client: h1ClientInsecure, wantProto: "HTTP/1.1"},
		{name: "HTTP/1.1 - Without CA Certs", client: h1Client, wantClientErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest("GET", testURL, nil)
			require.NoError(t, err)
			resp, err := tt.client.Do(req)
			if tt.wantClientErr {
				// we expect an error, exit the test
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// in this test i always expect 200
			require.Equal(t, http.StatusOK, resp.StatusCode)

			// reading the body, decode the body into result

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			result := new(serverResponse)

			err = json.Unmarshal(body, result)
			require.NoError(t, err)

			// ensure that we got a secure url
			require.True(t, strings.HasPrefix(result.URL, "https://"))
			// ensure that client's proto is what we expect
			require.Equal(t, tt.wantProto, result.Proto)
		})
	}
}
