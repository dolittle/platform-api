#!/bin/sh
#
# 
#
#rm -rf /tmp/dolittle-local-dev
#mkdir /tmp/dolittle-local-dev
#rsync -rv --exclude='.git' ~/dolittle/git/Operations/ /tmp/dolittle-local-dev

echo "Exploring"
exit 1
DATA=$(kubectl get ns --no-headers -o custom-columns=":metadata.name" | grep '^application-')
DIR="/tmp/dolittle-k8s-diff"
rm -rf $DIR
mkdir $DIR

for NAMESPACE in $DATA; do
    echo $NAMESPACE
    continue
    OUTPUT="${DIR}/$NAMESPACE.log"
    CMD=$(cat <<_EOF_

GIT_REPO_BRANCH=dev \
GIT_REPO_DRY_RUN=true \
GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
GIT_REPO_DIRECTORY_ONLY=true \
go run main.go tools studio build-terraform-info 
_EOF_

#go run main.go tools studio build-application-info ${NAMESPACE} > ${OUTPUT}
)
    eval $CMD
done

