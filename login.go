package main

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func Login() string {
	randomState := uuid.New().String()

	loginAttempts[randomState] = time.Now().Add(time.Hour)

	externalLoginURL := oauthConfig.AuthCodeURL(randomState, oauth2.AccessTypeOffline)

	return externalLoginURL
}
