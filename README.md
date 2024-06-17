# httpbulb
Implementation of `httpbin` for `Go`.

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



