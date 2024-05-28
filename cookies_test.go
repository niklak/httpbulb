package httpbulb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CookiesSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *http.Client
}

func (s *CookiesSuite) SetupSuite() {

	handleFunc := NewRouter()
	s.testServer = httptest.NewServer(handleFunc)

	jar, _ := cookiejar.New(nil)

	s.client = &http.Client{Jar: jar}
}

func (s *CookiesSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *CookiesSuite) TestCookies() {
	type serverResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	apiURL := fmt.Sprintf("%s/cookies", s.testServer.URL)

	req, err := http.NewRequest("GET", apiURL, nil)
	assert.NoError(s.T(), err)

	req.Header.Set("Cookie", "k1=v1; k2=v2")

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	res := &serverResponse{}

	err = json.NewDecoder(resp.Body).Decode(res)
	assert.NoError(s.T(), err)

	expected := map[string][]string{"k1": {"v1"}, "k2": {"v2"}}

	assert.Equal(s.T(), expected, res.Cookies)

}

func (s *CookiesSuite) TestSetCookies() {
	type serverResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	apiU, err := url.Parse(s.testServer.URL)
	assert.NoError(s.T(), err)

	apiU = apiU.ResolveReference(
		&url.URL{Path: "/cookies/set", RawQuery: "k3=v"},
	)

	req, err := http.NewRequest("GET", apiU.String(), nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	res := &serverResponse{}

	err = json.NewDecoder(resp.Body).Decode(res)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "v", res.Cookies["k3"][0])

}

func (s *CookiesSuite) TestSetCookie() {
	type serverResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	apiU, err := url.Parse(s.testServer.URL)
	assert.NoError(s.T(), err)

	apiU = apiU.ResolveReference(&url.URL{Path: "/cookies/set/k4/v"})

	req, err := http.NewRequest("GET", apiU.String(), nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	res := &serverResponse{}

	err = json.NewDecoder(resp.Body).Decode(res)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "v", res.Cookies["k4"][0])

}

func (s *CookiesSuite) TestDeleteCookies() {
	type serverResponse struct {
		Cookies map[string][]string `json:"cookies"`
	}

	apiU, err := url.Parse(s.testServer.URL)
	assert.NoError(s.T(), err)

	apiU = apiU.ResolveReference(
		&url.URL{Path: "/cookies/delete", RawQuery: "k5=v"},
	)

	s.client.Jar.SetCookies(
		apiU,
		[]*http.Cookie{{Name: "k5", Value: "v", Path: "/"}},
	)

	req, err := http.NewRequest("GET", apiU.String(), nil)
	assert.NoError(s.T(), err)

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	res := &serverResponse{}

	err = json.NewDecoder(resp.Body).Decode(res)
	assert.NoError(s.T(), err)

	_, isPresent := res.Cookies["k5"]

	assert.False(s.T(), isPresent)

}

func TestCookiesSuite(t *testing.T) {
	suite.Run(t, new(CookiesSuite))
}
