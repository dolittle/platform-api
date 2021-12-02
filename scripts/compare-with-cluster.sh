#!/bin/sh

DATA=$(go run main.go tools get-microservices)

DATA=$(cat /tmp/microservices.ndjson | jq -c 'select(.file_name)| .')

DIR="/tmp/dolittle-k8s-diff"
rm -rf $DIR
mkdir $DIR

for i in $DATA; do
    FILE_NAME=$(echo $i | jq -r '.file_name')
    
    UNIQUE_ID=$(echo "${FILE_NAME}" | md5 | awk '{print $1}' )
    OUTPUT="${DIR}/${UNIQUE_ID}.diff"

    CMD=$(cat <<_EOF_
kubectl diff -f ${FILE_NAME} > ${OUTPUT}
_EOF_
)
    eval $CMD
    echo $?
done

