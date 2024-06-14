package httpbulb

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// DigestAuthHandle prompts the user for authorization using HTTP Digest Auth.
func DigestAuthHandle(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "user")
	passwd := chi.URLParam(r, "passwd")
	algorithm := chi.URLParam(r, "algorithm")
	qop := chi.URLParam(r, "qop")
	staleAfter := chi.URLParam(r, "stale_after")

	requireCookieParam := strings.ToLower(r.URL.Query().Get("require-cookie"))

	secureCookie := getURLScheme(r) == schemeHttps

	var requireCookie bool

	switch requireCookieParam {
	case "true", "1", "t":
		requireCookie = true
	}

	switch algorithm {
	case "MD5", "SHA-256", "SHA-512":
	default:
		algorithm = "MD5"
	}

	if staleAfter == "" {
		staleAfter = "never"
	}

	_, hasCookie := r.Header["Cookie"]

	authorization := r.Header.Get("Authorization")

	credentials, err := parseDigestAuth(authorization)

	if err != nil || (requireCookie && !hasCookie) {
		// 401 response
		setCookie(w, "stale_after", staleAfter, secureCookie)
		setCookie(w, "fake", "fake_value", secureCookie)
		writeDigestChallengeResponse(w, r, "httpbulb", qop, algorithm, false)
		return
	}

	if requireCookie && getCookie(r, "fake") != "fake_value" {
		// 403 response
		setCookie(w, "fake", "fake_value", secureCookie)
		JsonError(w, "missing cookie set on challenge", http.StatusForbidden)
		return
	}

	currentNonce := credentials["nonce"]

	staleAfterValue := getCookie(r, "stale_after")

	lastNonce := getCookie(r, "last_nonce")

	if (lastNonce != "" && currentNonce == lastNonce) || staleAfterValue == "0" {
		setCookie(w, "stale_after", staleAfter, secureCookie)
		setCookie(w, "fake", "fake_value", secureCookie)
		setCookie(w, "last_nonce", currentNonce, secureCookie)
		writeDigestChallengeResponse(w, r, "httpbulb", qop, algorithm, true)
		return
	}

	if !checkDigestAuth(r, credentials, user, passwd) {

		setCookie(w, "stale_after", staleAfter, secureCookie)
		setCookie(w, "fake", "fake_value", secureCookie)
		setCookie(w, "last_nonce", currentNonce, secureCookie)
		writeDigestChallengeResponse(w, r, "httpbulb", qop, algorithm, false)
		return
	}

	if staleAfterValue != "" {
		setCookie(w, "stale_after", nextStaleAfterValue(staleAfterValue), secureCookie)
	}
	setCookie(w, "fake", "fake_value", secureCookie)
	writeJsonResponse(w, http.StatusOK, AuthResponse{Authenticated: true, User: user})

}

func parseDigestAuth(authHeader string) (credentials map[string]string, err error) {

	if authHeader == "" {
		err = fmt.Errorf("missing Authorization header")
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	authType := parts[0]
	authInfo := parts[1]

	if strings.ToLower(authType) != "digest" {
		err = fmt.Errorf("supported authorization type is Digest")
		return
	}

	credentials = parseHeaderValues(authInfo)
	requiredCredentials := []string{"username", "realm", "nonce", "uri", "response"}

	for _, cred := range requiredCredentials {
		if _, ok := credentials[cred]; !ok {
			err = fmt.Errorf("missing required credential %s", cred)
			return
		}
	}

	if _, hasQop := credentials["qop"]; !hasQop {
		_, hasNC := credentials["nc"]
		_, hasCNonce := credentials["cnonce"]

		if !hasNC || !hasCNonce {
			err = fmt.Errorf("missing required credentials nc and cnonce")
			return
		}
	}

	return
}

func parseHeaderValues(value string) map[string]string {

	parts := strings.Split(value, ",")
	m := make(map[string]string)

	for _, part := range parts {
		part := strings.TrimSpace(part)
		p := strings.SplitN(part, "=", 2)
		key := p[0]
		value := strings.Trim(p[1], `"`)
		m[key] = value
	}
	return m
}

func writeDigestChallengeResponse(w http.ResponseWriter, r *http.Request, realm, qop, algorithm string, stale bool) {
	ts := time.Now().Unix()

	b := make([]byte, 10)
	rand.Read(b)

	opaqueB := make([]byte, 10)
	rand.Read(opaqueB)

	nonceBuf := new(bytes.Buffer)
	nonceBuf.WriteString(r.RemoteAddr)
	nonceBuf.WriteString(":")
	nonceBuf.WriteString(fmt.Sprintf("%d", ts))
	nonceBuf.WriteString(":")
	nonceBuf.Write(b)

	nonce := hash(nonceBuf.Bytes(), algorithm)
	opaque := hash(opaqueB, algorithm)

	if qop == "" {
		qop = "auth"
	}

	value := fmt.Sprintf(
		"Digest qop=%s, realm=%s, algorithm=%s, nonce=%s, opaque=%s stale=%t",
		qop, realm, algorithm, nonce, opaque, stale)
	w.Header().Set("WWW-Authenticate", value)
	w.WriteHeader(http.StatusUnauthorized)

}

func hash(data []byte, algorithm string) string {
	var h []byte

	switch algorithm {
	case "SHA-256":
		checksumB := sha256.Sum256(data)
		h = checksumB[:]
	case "SHA-512":
		checksumB := sha512.Sum512(data)
		h = checksumB[:]
	default:
		checksumB := md5.Sum(data)
		h = checksumB[:]
	}

	return hex.EncodeToString(h)
}

func ha1(realm, username, password, algorithm string) string {
	a1 := []byte(fmt.Sprintf("%s:%s:%s", username, realm, password))
	return hash(a1, algorithm)
}

func ha2(method, uri, algorithm string) string {
	a2 := []byte(fmt.Sprintf("%s:%s", method, uri))
	return hash(a2, algorithm)
}

func compileDigestResponse(credentials map[string]string, password, method, uri string) string {
	algorithm := credentials["algorithm"]
	qop := credentials["qop"]
	nonce := credentials["nonce"]
	nc := credentials["nc"]
	cnonce := credentials["cnonce"]

	ha1Value := ha1(credentials["realm"], credentials["username"], password, algorithm)
	ha2Value := ha2(method, uri, algorithm)

	var resp string
	switch qop {
	case "auth", "auth-int":
		resp = fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1Value, nonce, nc, cnonce, qop, ha2Value)

	default:
		// actually qop must be either `auth` or `auth-int`
		// TODO: remove support for custom
		resp = fmt.Sprintf("%s:%s:%s", ha1Value, nonce, ha2Value)
	}

	return hash([]byte(resp), algorithm)
}

func checkDigestAuth(r *http.Request, credentials map[string]string, username, password string) (ok bool) {

	if credentials == nil {
		return
	}
	if credentials["username"] != username {
		return
	}

	responseHash := compileDigestResponse(credentials, password, r.Method, getAbsoluteURL(r))
	expectedHash := credentials["response"]
	if subtle.ConstantTimeCompare([]byte(responseHash), []byte(expectedHash)) == 1 {
		ok = true
	}
	return
}

func nextStaleAfterValue(staleAfter string) string {
	if num, err := strconv.Atoi(staleAfter); err == nil {
		return fmt.Sprintf("%d", num-1)
	}
	return "never"

}

// TODO: clean up this file
