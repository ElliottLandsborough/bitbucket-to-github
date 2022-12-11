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

	c := getRepositories(httpClient, "bitbucket")

	cloneRepositories(c, "/tmp/foo/")

	// Create new gh oauth
	var gh GitHubOauth

	// Open a browser if possible or echo the url to command line
	go OpenBrowser(gh.generateOauthUrl(gitHubClientID))

	handlePosix()
}
