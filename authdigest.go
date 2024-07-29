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

type digestCredentials struct {
	username  string
	realm     string
	nonce     string
	uri       string
	response  string
	qop       string
	nc        string
	cnonce    string
	algorithm string
}

func (d *digestCredentials) fromMap(m map[string]string) *digestCredentials {

	for k, v := range m {
		switch k {
		case "username":
			d.username = v
		case "realm":
			d.realm = v
		case "nonce":
			d.nonce = v
		case "uri":
			d.uri = v
		case "response":
			d.response = v
		case "qop":
			d.qop = v
		case "nc":
			d.nc = v
		case "cnonce":
			d.cnonce = v
		case "algorithm":
			d.algorithm = v
		}
	}
	return d

}

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

	// Actually `algorithm` path parameter is not relevant,
	// because algorithm will be taken from Authorization header.
	// It will be relevant only if the Authorization header is not set or is not Digest.

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

	currentNonce := credentials.nonce

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

func parseDigestAuth(authHeader string) (dig *digestCredentials, err error) {

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

	cm := parseHeaderValues(authInfo)
	requiredCredentials := []string{"username", "realm", "nonce", "uri", "response"}

	for _, cred := range requiredCredentials {
		if _, ok := cm[cred]; !ok {
			err = fmt.Errorf("missing required credential %q", cred)
			return
		}
	}

	dig = (&digestCredentials{}).fromMap(cm)

	if dig.qop != "" {
		if dig.nc == "" || dig.cnonce == "" {
			err = fmt.Errorf("missing required credentials 'nc' and 'cnonce'")
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

	nonce := hexDigest(nonceBuf.Bytes(), algorithm)
	opaque := hexDigest(opaqueB, algorithm)

	if qop == "" {
		qop = "auth"
	}

	value := fmt.Sprintf(
		"Digest qop=%s, realm=%s, algorithm=%s, nonce=%s, opaque=%s stale=%t",
		qop, realm, algorithm, nonce, opaque, stale)
	w.Header().Set("WWW-Authenticate", value)
	w.WriteHeader(http.StatusUnauthorized)

}

func hexDigest(data []byte, algorithm string) string {
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
	return hexDigest(a1, algorithm)
}

func ha2(method, uri, algorithm string) string {
	a2 := []byte(fmt.Sprintf("%s:%s", method, uri))
	return hexDigest(a2, algorithm)
}

func compileDigestResponse(dig *digestCredentials, password, method, uri string) string {

	ha1Value := ha1(dig.realm, dig.username, password, dig.algorithm)
	ha2Value := ha2(method, uri, dig.algorithm)

	var resp string
	switch dig.qop {
	case "auth", "auth-int":
		resp = fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1Value, dig.nonce, dig.nc, dig.cnonce, dig.qop, ha2Value)

	default:
		resp = fmt.Sprintf("%s:%s:%s", ha1Value, dig.nonce, ha2Value)
	}

	return hexDigest([]byte(resp), dig.algorithm)
}

func checkDigestAuth(r *http.Request, dig *digestCredentials, username, password string) (ok bool) {

	if dig == nil {
		return
	}
	if dig.username != username {
		return
	}

	responseHexDigest := compileDigestResponse(dig, password, r.Method, r.RequestURI)
	expectedHexDigest := dig.response
	if subtle.ConstantTimeCompare([]byte(responseHexDigest), []byte(expectedHexDigest)) == 1 {
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
