package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

var GITHUB_TOKEN string
var BITBUCKET_TOKEN string

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

type GitHubOauth struct {
}

func (r *GitHubOauth) generateOauthUrl(clientID string) string {
	url1 := "https://github.com/login/oauth/authorize?client_id=" + clientID
	url2 := "&redirect_uri=http://localhost:8080/oauth/redirect/github"
	url3 := "&scope=repo"

	url := fmt.Sprintf("%s%s%s", url1, url2, url3)

	return url
}

type BitBucketOauth struct {
}

func (r *BitBucketOauth) generateOauthUrl(clientID string) string {
	url1 := "https://bitbucket.org/site/oauth2/authorize?client_id=" + clientID
	url2 := "&redirect_uri=http://localhost:8080/oauth/redirect/bitbucket"
	url3 := "&response_type=code&scope=repository"

	url := fmt.Sprintf("%s%s%s", url1, url2, url3)

	return url
}

func getToken(httpClient http.Client, w http.ResponseWriter, r *http.Request, clientID string, clientSecret string, provider string) {
	// First, we need to get the value of the `code` query param
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not parse query: %v\n", err)
		fmt.Fprintf(os.Stdout, "could not parse query: %v\n", err)
		os.Exit(1)
	}
	code := r.FormValue("code")

	// github only
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)
	body := bytes.NewBufferString("")

	// bitbucket only
	if provider == "bitbucket" {
		data := url.Values{}
		data.Set("grant_type", "client_credentials")
		data.Set("code", code)
		reqURL = "https://bitbucket.org/site/oauth2/access_token"
		body = bytes.NewBufferString(data.Encode())
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, body) //
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not create HTTP request: %v\n", err)
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v\n", err)
		os.Exit(1)
	}

	// as JSON
	req.Header.Set("accept", "application/json")

	// Some bitbucket specific headers and auth
	if provider == "bitbucket" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(clientID, clientSecret)
	}

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not send HTTP request: %v\n", err)
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	// Empty struct to parse the json into
	var t OAuthAccessResponse

	// Parse the request body into the `OAuthAccessResponse` struct
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not parse JSON response: %v\n", err)
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v\n", err)
		os.Exit(1)
	}

	switch provider {
	case "github":
		GITHUB_TOKEN = t.AccessToken
	case "bitbucket":
		BITBUCKET_TOKEN = t.AccessToken
	}

	fmt.Fprintf(os.Stdout, "Access token: %s\n", t.AccessToken)

	fmt.Fprintf(w, "Success. You can close this tab.\n")
}

func getGlobToken(provider string) string {
	var t string

	switch provider {
	case "github":
		t = GITHUB_TOKEN
	case "bitbucket":
		t = BITBUCKET_TOKEN
	}

	return t
}

// Waits for oauth access response before continuing (todo? Replace with channel)
func waitForOAuthAccessResponse(provider string) {

	for {
		if len(getGlobToken(provider)) > 0 {
			break
		}
	}
}
