# Port-forward to nats
```sh
kill -9 $(lsof -ti tcp:4222)
kubectl -n application-11b6cf47-5d9f-438f-8116-0d9828654657 port-forward svc/nats 4222:4222 &
```

# Run the server
## With Nats
```sh
WEBHOOK_REPO="nats" NATS_SERVER="127.0.0.1" STAN_CLUSTER_ID="stan" STAN_CLIENT_ID="webhook-1" go run main.go raw-data-log server
```
## With stdout
```sh
WEBHOOK_REPO="stdout" go run main.go raw-data-log server
```

# Post some content

```sh
curl -XPOST \
'localhost:8080/webhook/yo/yoyo' \
-d '
{
  "id": "11b6cf47-5d9f-438f-8116-0d9828654657",
  "name": "Taco",
  "tenantId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3"
}'
```

# Setup in k8s
- hardcoded to customer-chris
## Nats
```sh
go run main.go raw-data-log build-nats --kube-config="/Users/freshteapot/.kube/config" --action=upsert ./k8s/single-server-nats.yml
```

## Stan with in memory store
```sh
go run main.go raw-data-log build-nats --kube-config="/Users/freshteapot/.kube/config" --action=upsert ./k8s/single-server-stan-memory.yml
```


# Read from the logs
```sh
TOPIC=topic.todo \
STAN_CLIENT_ID=nats-reader \
STAN_CLUSTER_ID=stan \
NATS_SERVER=127.0.0.1 \
go run main.go raw-data-log read-logs
```


# Docker build
```sh
docker build -f ./Dockerfile -t dolittle/platform-api:dev-x .
docker tag dolittle/platform-api:dev-x 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/platform-api:dev-x
docker push 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/platform-api:dev-x
az acr login -n 508c17455f2a4b4cb7a52fbb1484346d
```


```sh
docker tag dolittle/platform-api:dev-x 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/platform-api:latest
docker push 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/platform-api:latest
```



```
curl -XPOST \
'https://freshteapot-taco.dolittle.cloud/webhook2/hello/chris' \
-d '
{
  "hello": "world"
}'


``

`