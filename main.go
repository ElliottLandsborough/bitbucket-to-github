package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/joho/godotenv"
)

var t OAuthAccessResponse

func main() {
	// Load .env file (see example)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the settings we need
	clientID := os.Getenv("BITBUCKET_CLIENT_ID")
	clientSecret := os.Getenv("BITBUCKET_SECRET")

	httpClient := http.Client{}

	// Create a new redirect route route
	http.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		getToken(httpClient, w, r, clientID, clientSecret)
		fmt.Fprintf(os.Stdout, "Access token: %s\n", t.AccessToken)
	})

	// Listen for connections
	go http.ListenAndServe(":8080", nil)

	// Open a browser if possible or echo the url to command line
	go OpenBrowser(generateOauthUrl(clientID))

	go waitForOAuthAccessResponse(httpClient)

	handlePosix()
}

func waitForOAuthAccessResponse(httpClient http.Client) {
	// todo? Replace with channel.
	// In this case a for is fine because only one  goroutine can change `t`
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
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

type GitHubRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	SshUrl      string `json:"ssh_url"`
}

type BitBucketResponse struct {
	Repos   []BitBucketRepo `json:"values"`
	Size    int             `json:"size"`
	Page    int             `json:"page"`
	Pagelen int             `json:"pagelen"`
	Next    string          `json:"next"`
}

type BitBucketRepo struct {
	Slug        string `json:"slug"` // git@bitbucket.org:[slug]
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
}

var browserCommands = map[string]string{
	"windows": "start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

func generateOauthUrl(clientID string) string {
	//url1 := "https://github.com/login/oauth/authorize?client_id=" + clientID
	url1 := "https://bitbucket.org/site/oauth2/authorize?client_id=" + clientID
	url2 := "&redirect_uri=http://localhost:8080/oauth/redirect"
	//url3 := "&scope=repo"
	url3 := "&response_type=code&scope=repository"

	url := fmt.Sprintf("%s%s%s", url1, url2, url3)

	return url
}

func OpenBrowser(uri string) {
	run, ok := browserCommands[runtime.GOOS]
	if !ok {
		fmt.Fprintf(os.Stdout, "don't know how to open things on %s platform\n", runtime.GOOS)
		fmt.Fprintf(os.Stdout, "Click this link to authorize repository access: %s\n", uri)
	}
	cmd := exec.Command(run, uri)
	cmd.Start()
}

func getToken(httpClient http.Client, w http.ResponseWriter, r *http.Request, clientID string, clientSecret string) {
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

	// github
	//reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)

	// bitbucket
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("code", code)
	reqURL := "https://bitbucket.org/site/oauth2/access_token"

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not create HTTP request: %v\n", err)
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v\n", err)
		os.Exit(1)
	}
	// We set this header since we want the response

	// bitbucket
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// as JSON
	req.Header.Set("accept", "application/json")
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

	for _, repo := range s.Repos {
		fmt.Fprintf(os.Stdout, "Name: %s\n", repo.Name)
	}

	return s
}

// https://gobyexample.com/signals
func handlePosix() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done
	fmt.Println("exiting")
}
