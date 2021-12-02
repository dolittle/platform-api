#!/bin/sh

export GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev"
export GIT_REPO_DIRECTORY_ONLY="false"  
export GIT_REPO_SSH_KEY="/Users/freshteapot/.ssh/dolittle_operations"
export GIT_REPO_BRANCH="main"
export GIT_REPO_URL="git@github.com:dolittle-platform/Operations.git"

DATA=$(go run main.go tools get-microservices)

for i in $DATA; do
    applicationID=$(echo $i | jq -r '.applicationId')
    microserviceID=$(echo $i | jq -r '.microserviceId')
    environment=$(echo $i | jq -r '.environment')
    
    output="/tmp/microservices.ndjson"
    touch ${output}
    cmd=$(cat <<_EOF_
go run main.go tools get-head-image  --application-id="${applicationID}" --environment="${environment}" --microservice-id="${microserviceID}" >> ${output}
_EOF_
)
    eval $cmd
    echo $?
done

