package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type GitHubResponse struct {
	Repos []GitHubRepo
}

type GitHubRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	SshUrl      string `json:"ssh_url"`
}

func (b *GitHubResponse) toClonables() map[string]Clonable {
	s := make(map[string]Clonable)

	repos := b.Repos

	for _, r := range repos {
		var c Clonable
		c.Name = r.Name
		c.SshUrl = r.SshUrl
		s[r.Name] = c
	}

	return s
}

type BitBucketResponse struct {
	Repos   []BitBucketRepo `json:"values"`
	Size    int             `json:"size"`
	Page    int             `json:"page"`
	Pagelen int             `json:"pagelen"`
	Next    string          `json:"next"`
}

func (b *BitBucketResponse) toClonables() map[string]Clonable {
	s := make(map[string]Clonable)

	repos := b.Repos

	for _, r := range repos {
		var c Clonable
		c.Name = r.Name
		c.SshUrl = r.cloneUrl()
		s[r.Name] = c
	}

	return s
}

type BitBucketRepo struct {
	Slug        string `json:"slug"` // git@bitbucket.org:[username]/[slug]
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
}

func (r *BitBucketRepo) cloneUrl() string {
	return "git@bitbucket.org:" + os.Getenv("BITBUCKET_USER") + "/" + r.Slug
}

func getRepositories(provider string) map[string]Clonable {
	waitForOAuthAccessResponse(provider)

	reqURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%v?pagelen=100", os.Getenv("BITBUCKET_USER"))

	if provider == "github" {
		reqURL = "https://api.github.com/user/repos?per_page=100"
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+getBearerToken(provider))

	if provider == "github" {
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	}

	httpClient := http.Client{}

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	var c map[string]Clonable

	if provider == "github" {
		var s GitHubResponse
		if err := json.NewDecoder(res.Body).Decode(&s.Repos); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Github repository count: %v\n", len(s.Repos))

		c = s.toClonables()
	}

	if provider == "bitbucket" {
		var s BitBucketResponse
		if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Bitbucket repository count: %v\n", len(s.Repos))
		fmt.Fprintf(os.Stdout, "SIZE: %v\n", s.Size)
		fmt.Fprintf(os.Stdout, "PAGELEN: %v\n", s.Pagelen)
		fmt.Fprintf(os.Stdout, "PAGE: %v\n", s.Page)

		c = s.toClonables()
	}

	return c
}

func gitHubRepoHasContributors(repo Clonable) bool {
	waitForOAuthAccessResponse("github")

	reqURL := fmt.Sprintf("https://api.github.com/repos/%v/%v/stats/contributors", os.Getenv("GITHUB_USER"), repo.Name)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+getBearerToken("github"))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	httpClient := http.Client{}

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
	}

	if len(string(b)) == 0 {
		return false
	}

	return true
}

func createPrivateGithubRepos(s map[string]Clonable) {
	for _, r := range s {
		createPrivateGithubRepo(r)
		break // todo: remove
	}
}

func createPrivateGithubRepo(c Clonable) {
	waitForOAuthAccessResponse("github")

	reqURL := "https://api.github.com/user/repos"

	values := map[string]string{"name": c.Name, "private": "true"}

	jsonValue, _ := json.Marshal(values)

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+getBearerToken("github"))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	httpClient := http.Client{}

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		fmt.Fprintf(os.Stdout, "ERROR\n")
		fmt.Fprintf(os.Stdout, "response Status: %v\n", res.Status)
		fmt.Fprintf(os.Stdout, "response Headers: %v\n", res.Header)
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Fprintf(os.Stdout, "response Body: %v\n", string(body))
	}

	fmt.Fprintf(os.Stdout, "Private github repo created: %v\n", c.Name)
}
