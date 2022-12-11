package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type GitHubRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	SshUrl      string `json:"ssh_url"`
}

func (r *GitHubRepo) cloneUrl() string {
	return r.SshUrl
}

type BitBucketResponse struct {
	Repos   []BitBucketRepo `json:"values"`
	Size    int             `json:"size"`
	Page    int             `json:"page"`
	Pagelen int             `json:"pagelen"`
	Next    string          `json:"next"`
}

func (b *BitBucketResponse) toClonables() []Clonable {
	var s []Clonable

	repos := b.Repos

	for _, r := range repos {
		var c Clonable
		c.Name = r.Name
		c.SshUrl = r.cloneUrl()
		s = append(s, c)
	}

	return s
}

type BitBucketRepo struct {
	Slug        string `json:"slug"` // git@bitbucket.org:[slug]
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
}

func (r *BitBucketRepo) cloneUrl() string {
	return "git@bitbucket.org:" + r.Slug
}

func getRepositories(httpClient http.Client) BitBucketResponse {
	bearer := "Bearer " + t.AccessToken
	//gitHubApiVersion := "2022-11-28" // github only
	bitbucketUserName := os.Getenv("BITBUCKET_USER") // bitbucket only

	// Next, lets for the HTTP request to call the github oauth enpoint
	// to get our access token
	//reqURL := fmt.Sprintf("https://api.github.com/user/repos?per_page=100")
	reqURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%v?pagelen=100", bitbucketUserName)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v\n", err)
		os.Exit(1)
	}
	// We set this header since we want the response
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", bearer) // github
	//req.Header.Set("X-GitHub-Api-Version", gitHubApiVersion) // github

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	// Parse the request body into the `OAuthAccessResponse` struct
	//var s []GitHubRepo
	var s BitBucketResponse
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v\n", err)
		os.Exit(1)
	}

	return s
}
