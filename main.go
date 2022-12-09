package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID := os.Getenv("BITBUCKET_CLIENT_ID")
	clientSecret := os.Getenv("BITBUCKET_SECRET")

	url1 := "https://github.com/login/oauth/authorize?client_id=" + clientID
	url2 := "&redirect_uri=http://localhost:8080/oauth/redirect"
	url3 := "&scope=repo"

	url := fmt.Sprintf("%s%s%s", url1, url2, url3)

	OpenBrowser(url)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	// We will be using `httpClient` to make external HTTP requests later in our code
	httpClient := http.Client{}

	// Create a new redirect route route
	http.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		t := getToken(httpClient, w, r, clientID, clientSecret)

		fmt.Fprintf(os.Stdout, "Access token: %s", t.AccessToken)
	})

	http.ListenAndServe(":8080", nil)
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

var browserCommands = map[string]string{
	"windows": "start",
	"darwin":  "open",
	"linux":   "xdg-open",
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

func getToken(httpClient http.Client, w http.ResponseWriter, r *http.Request, clientID string, clientSecret string) OAuthAccessResponse {
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
	var t OAuthAccessResponse
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not parse JSON response: %v", err)
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(w, "Success. You can close this tab.")

	return t
}
