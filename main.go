package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		fmt.Fprintf(os.Stdout, "Access token: %s", t.AccessToken)
	})

	// Listen for connections
	go http.ListenAndServe(":8080", nil)

	// Open a browser if possible or echo the url to command line
	go OpenBrowser(generateOauthUrl(clientID))

	go waitForOAuthAccessResponse()

	handlePosix()
}

func waitForOAuthAccessResponse() {
	// todo? Replace with channel.
	// In this case a for is fine because only one  goroutine can change `t`
	for {
		if len(t.AccessToken) > 0 {
			break
		}
	}

	//g := getBitbucketRepositories(httpClient)
	//fmt.Fprintf(os.Stdout, "%v", g)
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

type GitHubReposResponse struct {
	Name   string `json:"name"`
	SshUrl string `json:"ssh_url"`
}

var browserCommands = map[string]string{
	"windows": "start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

func generateOauthUrl(clientID string) string {
	url1 := "https://github.com/login/oauth/authorize?client_id=" + clientID
	url2 := "&redirect_uri=http://localhost:8080/oauth/redirect"
	url3 := "&scope=repo"

	url := fmt.Sprintf("%s%s%s", url1, url2, url3)

	return url
}

func OpenBrowser(uri string) {
	run, ok := browserCommands[runtime.GOOS]
	if !ok {
		fmt.Fprintf(os.Stdout, "don't know how to open things on %s platform", runtime.GOOS)
		fmt.Fprintf(os.Stdout, "Click this link to authorize repository access: %s", uri)
	}
	cmd := exec.Command(run, uri)
	cmd.Start()
}

func getToken(httpClient http.Client, w http.ResponseWriter, r *http.Request, clientID string, clientSecret string) {
	// First, we need to get the value of the `code` query param
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not parse query: %v", err)
		fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
		os.Exit(1)
	}
	code := r.FormValue("code")

	// Next, lets for the HTTP request to call the github oauth enpoint
	// to get our access token
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not create HTTP request: %v", err)
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		os.Exit(1)
	}
	// We set this header since we want the response
	// as JSON
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not send HTTP request: %v", err)
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	// Parse the request body into the `OAuthAccessResponse` struct
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not parse JSON response: %v", err)
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(w, "Success. You can close this tab.")
}

func getBitbucketRepositories(httpClient http.Client) GitHubReposResponse {
	var bearer = "Bearer " + t.AccessToken
	var gitHubApiVersion = "2022-11-28"

	// Next, lets for the HTTP request to call the github oauth enpoint
	// to get our access token
	reqURL := fmt.Sprintf("https://api.github.com/user/repos?per_page=100")
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		os.Exit(1)
	}
	// We set this header since we want the response
	// as JSON
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", bearer)
	req.Header.Set("X-GitHub-Api-Version", gitHubApiVersion)

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	//fmt.Fprintf("%v", res.Body)

	// Parse the request body into the `OAuthAccessResponse` struct
	var g GitHubReposResponse
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		os.Exit(1)
	}

	return g
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
