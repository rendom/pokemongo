package pokemongo

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

const (
	loginURL          string = "https://sso.pokemon.com/sso/login?service=https%3A%2F%2Fsso.pokemon.com%2Fsso%2Foauth2.0%2FcallbackAuthorize"
	oAuthURL          string = "https://sso.pokemon.com/sso/oauth2.0/accessToken"
	oAuthClientSecret string = "w8ScCUXJQc6kXKw8FiOhd8Fixzht18Dq3PEVkUCP5ZPxtgyWsbTvWHFLm2wNY0JR"
	oAuthClientID     string = "mobile-app_pokemon-go"
	oAuthRedirectURI  string = "https://www.nianticlabs.com/pokemongo/error"

	// UserAgent containts the UA that will be used for all requests
	UserAgent = "niantic"

	// Regex Pattern for parsing out ticket key from location header
	ticketPattern = "\\?ticket=([^$]+)"

	// Regex Pattern for parsing out token from oauth response
	tokenPattern = "access_token=([^&]+)"
)

// jData contains some tokens?
type jData struct {
	Lt        string `json:"lt"`
	Execution string `json:"execution"`
}

// Login will auhtenticate to Pokemon Traingin Club and return token on success
func (api *Client) Login(username string, password string) (string, error) {
	jdata, cookies, err := getJdata()
	if err != nil {
		return "", err
	}

	ticket, err := getTicket(username, password, jdata, cookies)
	if err != nil {
		return "", err
	}

	token, err := authenticate(ticket)
	if err != nil {
		return "", err
	}
	return token, nil
}

func getTicket(username string, password string, jd jData, cookies []*http.Cookie) (string, error) {
	pf := url.Values{}
	pf.Add("lt", jd.Lt)
	pf.Add("execution", jd.Execution)
	pf.Add("_eventId", "submit")
	pf.Add("username", username)
	pf.Add("password", password)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(pf.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.AddCookie(cookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	location := resp.Header.Get("Location")

	if location == "" {
		return "", errors.New("Location is empty")
	}

	re, err := regexp.Compile(ticketPattern)
	if err != nil {
		return "", err
	}

	m := re.FindStringSubmatch(location)
	if len(m) < 1 || m[1] == "" {
		return "", errors.New("No ticket found")
	}

	return m[1], nil
}

// Authenticate authenticates to PTC
func authenticate(ticket string) (string, error) {
	pf := url.Values{}
	pf.Add("client_id", oAuthClientID)
	pf.Add("redirect_uri", oAuthRedirectURI)
	pf.Add("client_secret", oAuthClientSecret)
	pf.Add("grant_type", "refresh_token")
	pf.Add("code", ticket)

	req, err := http.NewRequest("POST", oAuthURL, bytes.NewBufferString(pf.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	re, err := regexp.Compile(tokenPattern)
	if err != nil {
		return "", err
	}

	m := re.FindStringSubmatch(string(body))
	if len(m) < 1 || m[1] == "" {
		return "", errors.New("No token found")
	}

	return m[1], nil
}

func getJdata() (jData, []*http.Cookie, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return jData{}, nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return jData{}, nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jData{}, nil, err
	}

	var jd jData
	err = json.Unmarshal(body, &jd)
	if err != nil {
		return jData{}, nil, err
	}

	return jd, resp.Cookies(), nil
}
