# Run server
```sh
LISTEN_ON="localhost:8080" go run main.go microservice server --kube-config="/Users/freshteapot/.kube/config"
```

# Create microservice
## Base / Simple
```sh
curl -XPOST localhost:8080/microservice -d '
{
  "dolittle": {
    "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
    "tenantId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
    "microserviceId": "9f6a613f-d969-4938-a1ac-5b7df199bc40"
  },
  "name": "Order",
  "kind": "simple",
  "environment": "Dev",
  "extra": {
    "headImage": "453e04a74f9d42f2b36cd51fa2c83fa3.azurecr.io/taco/order:1.0.6",
    "runtimeImage": "dolittle/runtime:5.6.0",
    "ingress": {
      "path": "/",
      "pathType": "Prefix"
    }
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
    "tenantId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
    "microserviceId": "9f6a613f-d969-4938-a1ac-5b7df199bc41"
  },
  "name": "Webhook-101",
  "kind": "buisness-moments-adaptor",
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
```sh
curl -XDELETE localhost:8080/application/11b6cf47-5d9f-438f-8116-0d9828654657/microservice/9f6a613f-d969-4938-a1ac-5b7df199bc40
```

# Not done
## Create application
```sh
curl -XPOST localhost:8080/application -d '
{
  "id": "11b6cf47-5d9f-438f-8116-0d9828654657",
  "name": "Taco",
  "tenantId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3"
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
