# Run server
```sh
GIT_BRANCH="auto-dev" \
LISTEN_ON="localhost:8080" \
AZURE_SUBSCRIPTION_ID="XXX" \
go run main.go microservice server --kube-config="/Users/freshteapot/.kube/config"
```

# Build customers.json
```sh
cat azure.json | jq -r '[.|to_entries | .[] | select(.value.value.customer).value.value.customer] | unique_by(.guid)'
```

# Create microservice
## Base / Simple
```sh
curl -XPOST localhost:8081/microservice \
-H 'x-shared-secret: FAKE' \
-H 'Tenant-ID: 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3' \
-H 'User-ID: local-dev' \
-H "Content-Type: application/json" \
-d '
{
  "dolittle": {
    "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
    "customerId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
    "microserviceId": "c96ab893-7df5-4065-a105-930c2c264ac2"
  },
  "name": "Order1",
  "kind": "simple",
  "environment": "Dev",
  "extra": {
    "headImage": "nginxdemos/hello:latest",
    "runtimeImage": "dolittle/runtime:7.6.0",
    "ingress": {
      "path": "/",
      "pathType": "Prefix"
    },
    "isPublic": true
  }
}'
```

## Business Moments Adaptor

{
            id: 'm3-webhook-1-basic',
            name: 'M3 Webhook Connector Basic',
            kind: 'webhook',
            config: {
                domain: '',
                uriPrefix: '',
                kind: 'basic',
                config: {
                    username: 'iamtest1',
                    password: 'test123'
                }
            }
        },

## Name cant have spaces
```sh
curl -XPOST localhost:8080/microservice -d '
{
  "dolittle": {
    "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
    "customerId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
    "microserviceId": "9f6a613f-d969-4938-a1ac-5b7df199bc41"
  },
  "name": "Webhook-101",
  "kind": "business-moments-adaptor",
  "environment": "Dev",
  "extra": {
    "headImage": "453e04a74f9d42f2b36cd51fa2c83fa3.azurecr.io/businessmomentsadaptor:latest",
    "runtimeImage": "dolittle/runtime:5.6.0",
    "ingress": {
      "path": "/api/webhooks-ingestor",
      "pathType": "Prefix"
    },
    "connector": {
      "kind": "webhook",
      "config": {
          "kind": "basic",
          "config": {
              "username": "m3",
              "password": "johncarmack"
          }
      }
    }
  }
}'
```

# Delete microservice

The `environment` URL parameter is case sensitive.

```sh
curl -X DELETE -i localhost:8081/application/11b6cf47-5d9f-438f-8116-0d9828654657/environment/Dev/microservice/9f6a613f-d969-4938-a1ac-5b7df199bc40 \
-H 'x-shared-secret: FAKE' \
-H 'Tenant-ID: 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3' \
-H 'User-ID: local-dev'
```

# Not done
## Create application
```sh
curl -XPOST localhost:8080/application -d '
{
  "id": "11b6cf47-5d9f-438f-8116-0d9828654657",
  "name": "Taco",
  "customerId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3"
}'
```
## Create application environment
```sh
curl -XPOST localhost:8080/application/11b6cf47-5d9f-438f-8116-0d9828654657/environment -d '
{
  "name": "Dev",
  "domainPrefix": "freshteapot-taco"
}'
```


## Create tenant
```sh
curl -XPOST localhost:8080/tenant -d '
{
  "id": "TODO",
  "name": "TODO"
}'
```





curl -XPOST https://freshteapot-taco.dolittle.cloud/webhook/api/ -d '{"hello": "world"}'



curl -XPOST localhost:8080/microservice \
-H 'Content-Type: application/json' \
-H 'x-shared-secret: change' \
-H 'Tenant-ID: 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3' \
-H 'User-ID: be194a45-24b4-4911-9c8d-37125d132b0b' \
-d '
{
  "dolittle": {
    "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
    "customerId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
    "microserviceId": "8d7be4a0-7fcb-dc48-ab2b-28fd8dbdbd3b"
  },
  "name": "RawDataLogIngestor",
  "kind": "raw-data-log-ingestor",
  "environment": "Dev",
  "extra": {
    "headImage": "453e04a74f9d42f2b36cd51fa2c83fa3.azurecr.io/dolittle/platform/platform-api:dev-x",
    "runtimeImage": "dolittle/runtime:5.6.0",
    "ingress": {
      "path": "/api/webhooks",
      "pathType": "Prefix",
      "host": "freshteapot-taco.dolittle.cloud",
      "domainPrefix": "freshteapot-taco"
    },
    "webhooks": [
      {
        "authorization": "todo auth",
        "uriSuffix": "abc/abc",
        "kind": "abc/abc"
      }
    ],
    "webhookStatsAuthorization": "test stats"
  }
}
'

```sh
curl -XDELETE \
-H 'Content-Type: application/json' \
-H 'x-shared-secret: change' \
-H 'Tenant-ID: 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3' \
-H 'User-ID: be194a45-24b4-4911-9c8d-37125d132b0b' \
"localhost:8080/application/11b6cf47-5d9f-438f-8116-0d9828654657/environment/dev/microservice/8d7be4a0-7fcb-dc48-ab2b-28fd8dbdbd3b"
```
