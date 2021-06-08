

# Port-forward to nats
```sh
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
