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
	basePath := os.Getenv("BASEPATH")

	httpClient := http.Client{}

	// Bitbucket route
	http.HandleFunc("/oauth/redirect/bitbucket", func(w http.ResponseWriter, r *http.Request) {
		getToken(httpClient, w, r, bitBucketClientID, bitBucketClientSecret, "bitbucket")
	})

	// GitHub route
	http.HandleFunc("/oauth/redirect/github", func(w http.ResponseWriter, r *http.Request) {
		getToken(httpClient, w, r, gitHubClientID, gitHubClientSecret, "github")
	})

	// Listen for connections
	go http.ListenAndServe(":8080", nil)

	// Create new bb oauth
	var bb BitBucketOauth

	// Open a browser if possible or echo the url to command line
	go OpenBrowser(bb.generateOauthUrl(bitBucketClientID))

	bitBucketClonables := getRepositories("bitbucket")

	// Create new gh oauth
	var gh GitHubOauth

	// Open a browser if possible or echo the url to command line
	go OpenBrowser(gh.generateOauthUrl(gitHubClientID))

	githubClonables := getRepositories("github")

	// pushables = list of github repos that we want to push
	pushables, duplicates := GetPushableAndDuplicateRepos(bitBucketClonables, githubClonables)

	for key := range duplicates {
		fmt.Fprintf(os.Stdout, "Duplicate will not be pushed: %v\n", key)
	}

	// Only clone the ones we want to push
	cloneRepositories(pushables, basePath, "bitbucket")

	pushLocalReposToGithub(pushables, basePath)

	fmt.Fprintf(os.Stdout, "Finished\n")

	handlePosix()
}
