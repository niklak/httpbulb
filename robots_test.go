package httpbulb

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RobotsTxt(t *testing.T) {

	handleFunc := NewRouter()
	// Start a test server that will act as a proxy
	testServer := httptest.NewServer(handleFunc)

	defer testServer.Close()

	// Create an http client with a proxy

	testUrl := fmt.Sprintf("%s/robots.txt", testServer.URL)

	req, err := http.NewRequest("GET", testUrl, nil)
	assert.NoError(t, err)

	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	expectedBody := []byte(`
User-agent: *
Disallow: /deny
	`)

	assert.Equal(t, expectedBody, body)

}
