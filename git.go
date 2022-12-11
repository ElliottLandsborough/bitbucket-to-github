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

func cloneRepositories(s []Clonable, basePath string) {
	for _, repo := range s {
		fmt.Fprintf(os.Stdout, "Cloning: %s\n", repo.Name)
		fmt.Fprintf(os.Stdout, "git clone https://github.com/go-git/go-git: %s\n", repo.SshUrl)

		cloneDir := basePath + repo.Name

		if _, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
			URL:      "https://github.com/go-git/go-git",
			Progress: os.Stdout,
		}); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v\n", err)
			os.Exit(1)
		}

		return

		/*
		  echo "* $repo cloned, now creating on github..."
		  echo
		  #curl -u $GH_USERNAME:$GH_PASSWORD https://api.github.com/orgs/$GH_ORG/repos -d "{\"name\": \"$repo\", \"private\": true}"
		  echo
		  echo "* mirroring $repo to github..."
		  echo
		  #git push --mirror git@github.com:$GH_ORG/$repo.git && \
		  #  bb delete -u $BB_USERNAME -p $BB_PASSWORD --owner $BB_ORG $repo
		*/
	}
}
