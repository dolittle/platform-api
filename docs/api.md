

# Get from git
```
curl -XGET \
-H 'X-Shared-Secret: TODO-1' \
'localhost:8080/application/11b6cf47-5d9f-438f-8116-0d9828654657/microservices' | jq
```

# Live from the cluster

# Get applications by tenant
```sh
curl -XGET localhost:8080/live/tenant/453e04a7-4f9d-42f2-b36c-d51fa2c83fa3/applications
```

# Get microservices by application
```sh
curl -XGET "localhost:8080/live/application/fe7736bb-57fc-4166-bb91-6954f4dd4eb7/microservices" | jq
```

# Get pod status for a microservice
```sh
curl -XGET "localhost:8080/live/application/11b6cf47-5d9f-438f-8116-0d9828654657/microservice/9f6a613f-d969-4938-a1ac-5b7df199bc40/podstatus/dev" | jq
```

# Get logs from a pod
```sh
curl -XGET "localhost:8080/live/application/11b6cf47-5d9f-438f-8116-0d9828654657/pod/dev-order-846fbc7776-x79r/logs" | jq
```


# BusinessMoments
## BusinessMoment
### Post
```json
{
  "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
  "environment": "Dev",
  "microserviceId": "55adce7e-0aea-0346-bdfd-8ef06b91a953",
  "moment": {
    "name": "I am name",
    "uuid": "fake-uuid-123",
    "embeddingCode": "embedding func",
    "projectionCode": "projection func",
    "entityTypeId": "fake-entity-1234"
  }
}
```

## Entity
### Post
```
{
  "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
  "environment": "Dev",
  "microserviceId": "55adce7e-0aea-0346-bdfd-8ef06b91a953",
  "entity": {
    "entityTypeId": "fake-entity-123",
    "idNameForRetrival": "id",
    "name": "I am name",
    "filterCode": "filter func",
    "transformCode": "transform func"
  }
}
```


# Configmaps
```sh
curl -s -XGET \
-H 'Content-Type: application/json' \
-H 'x-shared-secret: change' \
-H 'Tenant-ID: 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3' \
-H 'User-ID: be194a45-24b4-4911-9c8d-37125d132b0b' \
localhost:8080/application/11b6cf47-5d9f-438f-8116-0d9828654657/configmap/hi | jq
```