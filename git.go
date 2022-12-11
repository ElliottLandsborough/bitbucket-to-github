package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

type Clonable struct {
	Name   string
	SshUrl string
}

func cloneRepositories(s map[string]Clonable, basePath string, provider string) {
	waitForOAuthAccessResponse(provider)

	for _, repo := range s {
		fmt.Fprintf(os.Stdout, "Cloning: %s\n", repo.Name)
		fmt.Fprintf(os.Stdout, "git clone: %s\n", repo.SshUrl)

		cloneDir := basePath + "/" + repo.Name

		if _, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
			URL:      repo.SshUrl,
			Progress: os.Stdout,
		}); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v\n", err)
			os.Exit(1)
		}
	}
}

func GetPushableAndDuplicateRepos(bitBucketClonables map[string]Clonable, githubClonables map[string]Clonable) (map[string]Clonable, map[string]Clonable) {
	pushables := make(map[string]Clonable)
	duplicates := make(map[string]Clonable)

	for key, c := range bitBucketClonables {
		// If same repo name is on github
		if _, ok := githubClonables[key]; ok {
			// And it has contributors
			if gitHubRepoHasContributors(c) {
				// We don't want to push it
				duplicates[key] = c

				continue
			}
		}

		// Otherwise, we do want to push it
		pushables[key] = c
	}

	return pushables, duplicates
}

func pushLocalReposToGithub(s map[string]Clonable, basePath string) {
	for _, r := range s {
		pushLocalRepoToGithub(r, basePath)
	}
}

func pushLocalRepoToGithub(c Clonable, basePath string) {
	path := basePath + "/" + c.Name

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stdout, "Path does not exist `%v`.\n", path)
		os.Exit(1)
	}

	r, err := git.PlainOpen(path)

	if err != nil {
		fmt.Fprintf(os.Stdout, "Could not open repository at `%v`. %v\n", path, err)
		os.Exit(1)
	}

	remoteURL := "git@github.com:" + os.Getenv("GITHUB_USER") + "/" + c.Name

	// push using default options
	err = r.Push(&git.PushOptions{
		RemoteURL: remoteURL,
	})

	if err != nil {
		fmt.Fprintf(os.Stdout, "Could not push repository at `%v` to `%v`. %v\n", path, remoteURL, err)
		os.Exit(1)
	}
}
