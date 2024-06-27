# httpbulb
[![Go Reference](https://pkg.go.dev/badge/github.com/niklak/httpbulb.svg)](https://pkg.go.dev/github.com/niklak/httpbulb)
[![Go](https://github.com/niklak/httpbulb/actions/workflows/go.yml/badge.svg)](https://github.com/niklak/httpbulb/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/niklak/httpbulb)](https://goreportcard.com/report/github.com/niklak/httpbulb)
[![codecov](https://codecov.io/gh/niklak/httpbulb/graph/badge.svg?token=8GI1ZDHIH8)](https://codecov.io/gh/niklak/httpbulb)

[![Edit niklak/httpbulb/main](https://codesandbox.io/static/img/play-codesandbox.svg)](https://codesandbox.io/p/github/niklak/httpbulb/main?embed=1)

A tool for testing http client capabilities.

An implementation of `httpbin` for `Go`.

This is a http mux handler based on `go-chi`. 

It is useful for testing purposes. **Actually**, it was created to work with `httptest` package, because `httptest` package allows you to run http server for your tests without touching the external server.

> [!NOTE]
> This handler follows the original `httpbin` API but in some cases it is more strict.
## Requirements

Go 1.18

## Installation
`
go get -u github.com/niklak/httpbulb
`


## The main differences between `httpbin` and `httpbulb`

- `args`, `form`, `files` and `headers` fields are represented by `map[string][]string`.
- `/status/{code}` endpoint does not handle status codes lesser than 200 or greater than 599.
- `/cookies-list` -- a new endpoint that returns a cookie list (`[]http.Cookie`) in the same order as it was received and parsed on the go http server.
- `/images`, `/encoding/utf8`, `/html`, `/json`, `/xml` endpoints support `Range` requests.
- `/delete`, `/get`, `/patch`, `/post`, `/put` endpoints also return field `proto` which can help to detect HTTP protocol version in the client-server connection.


## Examples

The main approach is to use `httpbulb` with `httptest.Server`.

<details>

<summary>Testing client's http2 support</summary>

```go
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

```


</details>

It is also possible to use `httpbulb` as a web-server.
The binary can be built with from `github.com/niklak/httpbulb/cmd/bulb` or you can use docker image `ghcr.io/niklak/httpbulb:latest`.

<details>
<summary>Running a web-server using docker-compose</summary>

```yaml

name: httpbulb

services:
  web:
    image: ghcr.io/niklak/httpbulb:latest
    restart: unless-stopped
    volumes:
    # If you require an HTTPS server, you should link the directories with tls certificates.
      - "./data/certs:/certs"
    ports:
      # map the bulb port to your external port
      - :4443:8080
    environment:
      # you can set server address as HOST:PORT or just :PORT
      - SERVER_ADDR=:8080
      # If you require an HTTPS server, you should set SERVER_CERT_PATH and SERVER_KEY_PATH.
      # If you work with self-signed certificates, you need to ensure that `root CA` is installed on the requesting machine.
      # Or you need to mark your requests as insecure. (InsecureSkipVerify: true for Go, or --insecure flag for curl)
      # If server unable to load certificates, it will produce a warning, but start serving an HTTP server. 
      - SERVER_CERT_PATH=/certs/server-host.pem
      - SERVER_KEY_PATH=/certs/server-host-key.pem
      - SERVER_READ_TIMEOUT=120s
      - SERVER_WRITE_TIMEOUT=120s
```

After starting the server with `docker compose` its ready to accept requests.

```bash

# with no certificates
curl -v http://localhost:4443/get

# with self-signed certificates and installed root CA.
curl -v https://localhost:4443/get

# with self-signed certificates but without installed root CA on requesting machine.
curl -v --insecure https://localhost:4443/get

# with real certificates
curl -v https://example.com:4443/get

```

</details>


## Endpoints


| Route | Methods | Description |
|:------|---------|-------------|
|`/delete`<br> `/get`<br>`/patch`<br> `/post`<br> `/put`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`| These are basic endpoints. They return a response with request's common information. **Unlike the original `httpbin` implementation, this handler doesn't read the request body for `DELETE` and `GET` requests**. `args`, `files`, `form`, and `headers` are always represented by a map of string lists (slices). |
|`/basic-auth` |`GET`| Prompts the user for authorization using HTTP Basic Auth. Returns 401 if authorization is failed. |
|`/hidden-basic-auth` |`GET`| Prompts the user for authorization using HTTP Basic Auth. Returns 404 if authorization is failed. |
|`/digest-auth/{qop}/{user}/{passwd}`<br><br>`/digest-auth/{qop}/{user}/{passwd}/{algorithm}`<br><br>`/digest-auth/{qop}/{user}/{passwd}/{algorithm}/{stale_after}` |`GET`| Prompts the user for authorization using HTTP Digest Auth. Returns 401 or 403 if authorization is failed. |
| `/bearer` |`GET`| Prompts the user for authorization using bearer authentication. Returns 401 if authorization is failed. |
| `/status/{codes}` |`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`| Returns status code or random status code if more than one are given. **This handler does not handle status codes lesser than 200 or greater than 599.** |
|`/headers` |`GET`| Return the incoming request's HTTP headers. |
|`/ip` |`GET`| Returns the requester's IP Address. |
|`/user-agent` |`GET`| Return the incoming requests's User-Agent header. |
|`/cache`|`GET`| Returns a 304 if an If-Modified-Since header or If-None-Match is present. Returns the same as a `/get` otherwise.|
|`/cache/{value}`|`GET`|Sets a Cache-Control header for n seconds.|
|`/etag/{etag}`|`GET`|Assumes the resource has the given etag and responds to If-None-Match and If-Match headers appropriately.|
|`/response-headers`| `GET`<br>`POST`| Returns a set of response headers from the query string.|
|`/brotli`|`GET`|Returns Brotli-encoded data.|
|`/deflate`|`GET`|Returns Deflate-encoded data.|
|`/deny`|`GET`|Returns page denied by robots.txt rules.|
|`/encoding/utf8`|`GET`|Returns a UTF-8 encoded body.|
|`/gzip`|`GET`|Returns Gzip-encoded data.|
|`/html`|`GET`|Returns a simple HTML document.|
|`/json`|`GET`|Returns a simple JSON document.|
|`/robots.txt`|`GET`|Returns some robots.txt rules.|
|`/xml`|`GET`|Returns a simple XML document.|
|`/base64/{value}`|`GET`|Decodes base64url-encoded string.|
|`/bytes/{n}`|`GET`|Returns n random bytes generated with given seed.|
|`/delay/{delay}`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`|Returns a delayed response (max of 10 seconds).|
|`/drip`|`GET`|Drips data over a duration after an optional initial delay.|
|`/links/{n}/{offset}`|`GET`|Generates a page containing n links to other pages which do the same.|
|`/range/{numbytes}`|`GET`|Streams n random bytes generated with given seed, at given chunk size per packet. Supports `Accept-Ranges` and `Content-Range` headers.|
|`/stream-bytes/{n}`|`GET`|Streams n random bytes generated with given seed, at given chunk size per packet.|
|`/stream/{n}`|`GET`|Streams n json messages.|
|`/uuid`|`GET`| Returns a UUID4.|
|`/cookies`|`GET`|Returns cookie data.|
|`/cookies-list`|`GET`| **Returns a cookie list (`[]http.Cookie`) in the same order as it was received and parsed on the server.**|
|`/cookies/delete`|`GET`|Deletes cookie(s) as provided by the query string and redirects to cookie list.|
|`/cookies/set`|`GET`|Sets cookie(s) as provided by the query string and redirects to cookie list.|
|`/cookies/set/{name}/{value}`|`GET`|Sets a cookie and redirects to cookie list.|
|`/image`|`GET`|Returns a simple image of the type suggest by the Accept header. Also supports a `Range` requests|
|`/image/{format:svg\|png\|jpeg\|webp|avif}`|`GET`| Returns an image with the given format. If the `format` is not matched it returns 404|
|`/absolute-redirect/{n}`|`GET`| Absolutely 302 Redirects `n` times. `Location` header will be an absolute URL.|
|`/redirect-to`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`|302/3XX Redirects to the given URL. `url` parameter is required and `status` parameter is optional.|
|`/redirect/{n}`|`GET`| 302 Redirects n times. `Location` header will be an absolute if `absolute=true` was sent as a query parameter.|
|`/relative-redirect/{n}`|`GET`| Relatively 302 Redirects n times. `Location` header will be a relative URL.|
|`/anything`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`|Returns anything passed in request data.|
|`/anything/{anything}`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`|Returns anything passed in request data.|

