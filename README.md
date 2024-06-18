# httpbulb

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
|`/image/{format:svg\|png\|jpeg\|webp}`|`GET`| Returns an image with the given format. If the `format` is not matched it returns 404|
|`/absolute-redirect/{n}`|`GET`| Absolutely 302 Redirects `n` times. `Location` header will be an absolute URL.|
|`/redirect-to`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`|302/3XX Redirects to the given URL. `url` parameter is required and `status` parameter is optional.|
|`/redirect/{n}`|`GET`| 302 Redirects n times. `Location` header will be an absolute if `absolute=true` was sent as a query parameter.|
|`/relative-redirect/{n}`|`GET`| Relatively 302 Redirects n times. `Location` header will be a relative URL.|
|`/anything`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`|Returns anything passed in request data.|
|`/anything/{anything}`|`DELETE`<br>`GET`<br>`PATCH`<br>`POST`<br>`PUT`|Returns anything passed in request data.|

## The main differences between `httpbin` and `httpbulb`

- `args`, `form`, `files` and `headers` fields are represented by `map[string][]string`.
- `/status/{code}` endpoint does not handle status codes lesser than 200 or greater than 599.
- `/cookies-list` -- a new endpoint that returns a cookie list (`[]http.Cookie`) in the same order as it was received and parsed on the go http server.
- `/images`, `/encoding/utf8`, `/html`, `/json`, `/xml` endpoints support `Range` requests.
- `/delete`, `/get`, `/patch`, `/post`, `/put` endpoints also return field `proto` which can help to detect HTTP version of your http client.