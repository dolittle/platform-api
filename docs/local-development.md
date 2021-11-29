
# Developement without remote git
```sh
GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
GIT_REPO_DIRECTORY_ONLY="true" \
GIT_REPO_BRANCH=main \
LISTEN_ON="localhost:8080" \
HEADER_SECRET="FAKE" \
AZURE_SUBSCRIPTION_ID="e7220048-8a2c-4537-994b-6f9b320692d7" \
go run main.go microservice server --kube-config="/Users/freshteapot/.kube/config"
```


# Developement with remote git
```sh
GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
GIT_REPO_DIRECTORY_ONLY="false" \
GIT_REPO_SSH_KEY="/Users/freshteapot/.ssh/dolittle_operations" \
GIT_REPO_BRANCH=dev \
GIT_REPO_URL="git@github.com:dolittle-platform/Operations.git" \
LISTEN_ON="localhost:8080" \
HEADER_SECRET="FAKE" \
AZURE_SUBSCRIPTION_ID="e7220048-8a2c-4537-994b-6f9b320692d7" \
go run main.go microservice server --kube-config="/Users/freshteapot/.kube/config"
```