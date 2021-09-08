
'
{
  
  "name": "PurchaseOrderAPI",
  "kind": "purchase-order",
  "environment": "Dev",
  "extra": {
    "headImage": "TODO",
    "runtimeImage": "TODO",
    "webhooks": [],
    "rawDataLogID": "TODO"
  }
}
'

## purchase-order-api
```sh
curl -XPOST \
-H 'Content-Type: application/json' \
-H 'x-shared-secret: FAKE' \
-H 'Tenant-ID: 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3' \
-H 'User-ID: local-dev' \
localhost:8081/microservice -d '
{
  "dolittle": {
    "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
    "tenantId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
    "microserviceId": "042256a7-2ee1-46cf-950b-0d75e36ea624"
  },
  "name": "PurchaseOrderApi",
  "kind": "purchase-order-api",
  "environment": "Dev",
  "extra": {
    "headImage": "dolittle/integrations-m3-purchaseorders:2.2.3",
    "runtimeImage": "dolittle/runtime:6.1.0",
    "webhooks": [],
    "rawDataLogName": "rawdatalogingestor"
  }
}'
```