# bitbucket-to-github

Go here: `https://bitbucket.org/[username]/workspace/settings/api`

(or `https://github.com/settings/applications/new` for github)

Add a new consumer:

- Name: `anything`
- Callback URL: `http://localhost:8080/oauth/redirect`
- URL: `http://localhost:8080/`
- This is a private consumer `[x]`
- Permissions: `repository -> read`