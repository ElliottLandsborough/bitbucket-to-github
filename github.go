package main

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var oauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("BITBUCKET_CLIENT_ID"),
	ClientSecret: os.Getenv("BITBUCKET_SECRET"),
	Endpoint:     endpoints.GitHub,
	Scopes: []string{
		"repository",
	},
}
