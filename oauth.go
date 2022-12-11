package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

var t OAuthAccessResponse

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

	// Next, lets for the HTTP request to call the github oauth enpoint
	// to get our access token

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

	// bitbucket only
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// as JSON
	req.Header.Set("accept", "application/json")

	// Bitbucket only
	req.SetBasicAuth(clientID, clientSecret)

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not send HTTP request: %v\n", err)
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	// Parse the request body into the `OAuthAccessResponse` struct
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not parse JSON response: %v\n", err)
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(w, "Success. You can close this tab.\n")
}
