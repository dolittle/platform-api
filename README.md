# platform-api
What is a platform without an api? :)


# Production
## Updating customers.json
- you might need admin access in Cluster-Three
```sh
go run main.go rebuild-db --kube-config="/Users/freshteapot/.kube/config" --with-secrets | jq  > customers.json
```
### Dev
- Filter out the production environments
```sh
cat customers.json| jq '[.[] | del(.applications[]|select(.environment == "Prod"))]' > customers.dev.json
```
## Add to the respective config
Source/V3/Kubernetes/System/Data/Backups/Dev/download-backup-v1-config-files.yml

Source/V3/Kubernetes/System/Data/Backups/Prod/download-backup-v1-config-files.yml


## Update image
# Manual Update Image

```sh
docker build -f ./Dockerfile -t dolittle/downloads:dev-x .
```

```sh
docker tag dolittle/downloads:dev-x 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/downloads:dev-x
```

```sh
docker push 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/downloads:dev-x
```


# Development

# Rebuild customers.json
- you might need admin access in Cluster-Three
```sh
go run main.go rebuild-db --kube-config="/Users/freshteapot/.kube/config" --with-secrets | jq  > customers.json
```

# Run server
- You need a copy of customers.json

```sh
HEADER_SECRET=fake PATH_TO_DB=./customers.json go run main.go server
```

# Get Customers

```sh
curl -XGET \
-H'x-secret: fake' \
"localhost:8080/share/logs/customers"
```

# Get Applications

```sh
curl -XGET \
-H'x-secret: fake' \
"localhost:8080/share/logs/applications/Customer-Chris"
```

# Get Latest
## By Application
```sh
curl -XGET \
-H'x-secret: fake' \
"localhost:8080/share/logs/latest/by/app/453e04a7-4f9d-42f2-b36c-d51fa2c83fa3/Taco"
```

## By Domain
```sh
curl -XGET \
-H'x-secret: fake' \
"localhost:8080/share/logs/latest/freshteapot-taco.dolittle.cloud"
```
# Create link
```sh
curl -XPOST \
-H "Content-Type: application/json" \
-H'x-secret: fake' \
"localhost:8080/share/logs/link" -d '
{
  "tenant_id": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
  "application": "Taco",
  "environment": "Dev",
  "file_path": "/taco-dev-backup/mongo/2021-01-22_00-34-11.gz.mongodump"
}
'
```
