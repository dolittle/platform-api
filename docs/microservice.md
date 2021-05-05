# Create microservice
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
# Delete microservice
```sh
curl -XDELETE localhost:8080/application/11b6cf47-5d9f-438f-8116-0d9828654657/microservice/9f6a613f-d969-4938-a1ac-5b7df199bc40
```

# Not done
## Create application
```sh
curl -XPOST localhost:8080/application -d '
{
  "id": "TODO",
  "name": "TODO",
  "environment": "Dev"
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
