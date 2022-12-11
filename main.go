package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (see example)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the settings we need
	bitBucketClientID := os.Getenv("BITBUCKET_CLIENT_ID")
	bitBucketClientSecret := os.Getenv("BITBUCKET_SECRET")
	gitHubClientID := os.Getenv("GITHUB_CLIENT_ID")
	gitHubClientSecret := os.Getenv("GITHUB_SECRET")

	httpClient := http.Client{}

	// Bitbucket route
	http.HandleFunc("/oauth/redirect/bitbucket", func(w http.ResponseWriter, r *http.Request) {
		getToken(httpClient, w, r, bitBucketClientID, bitBucketClientSecret, "bitbucket")
		fmt.Fprintf(os.Stdout, "Access token: %s\n", t.AccessToken)
	})

	// GitHub route
	http.HandleFunc("/oauth/redirect/github", func(w http.ResponseWriter, r *http.Request) {
		getToken(httpClient, w, r, gitHubClientID, gitHubClientSecret, "github")
		fmt.Fprintf(os.Stdout, "Access token: %s\n", t.AccessToken)
	})

	// Listen for connections
	go http.ListenAndServe(":8080", nil)

	// Create new bb oauth
	var bb BitBucketOauth

	// Open a browser if possible or echo the url to command line
	go OpenBrowser(bb.generateOauthUrl(bitBucketClientID))

	// Waits for oauth access response and then does some things
	waitForOAuthAccessResponseThenClone(httpClient)

	// Create new gh oauth
	var gh GitHubOauth

	// Open a browser if possible or echo the url to command line
	go OpenBrowser(gh.generateOauthUrl(gitHubClientID))

	handlePosix()
}

func waitForOAuthAccessResponseThenClone(httpClient http.Client) {
	// todo? Replace with channel.
	// In this case a `for`` is fine because only one goroutine can change `t`
	for {
		if len(t.AccessToken) > 0 {
			break
		}
	}

	s := getRepositories(httpClient)

	fmt.Fprintf(os.Stdout, "Repository count: %v\n", len(s.Repos))
	fmt.Fprintf(os.Stdout, "SIZE: %v\n", s.Size)
	fmt.Fprintf(os.Stdout, "PAGELEN: %v\n", s.Pagelen)
	fmt.Fprintf(os.Stdout, "PAGE: %v\n", s.Page)

	c := s.toClonables()

	cloneRepositories(c, "/tmp/foo/")
}
