# bitbucket-to-github

## What?

1. Clones your bitbucket repos to `/tmp/foo`
2. Creates any that don't already exist on github as private repos
3. Pushes them all to github

## Why?

## How?

Go here: `https://bitbucket.org/[username]/workspace/settings/api`

Add a new consumer:

- Name: `anything you want`
- Callback URL: `http://localhost:8080/oauth/redirect/bitbucket`
- URL: `http://localhost:8080/`
- This is a private consumer `[x]`
- Permissions: `repository -> read`

Then go here: `https://github.com/settings/developers`

Create new oauth app:

- Name: `anything you want`
- Homepage URL: `http://localhost:8080/`
- Callback URL: `http://localhost:8080/oauth/redirect/github`

```bash
go run *.go
```