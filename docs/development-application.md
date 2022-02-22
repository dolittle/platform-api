## Create application
- Wrong customerId does something odd
```sh
curl -XPOST \
-H 'Content-Type: application/json' \
-H 'x-shared-secret: FAKE' \
-H 'Tenant-ID: 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3' \
-H 'User-ID: local-dev' \
localhost:8081/application -d '
{
  "id": "fake-application-123",
  "name": "Taco2",
  "customerId": "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
  "environments": ["Dev", "Test", "Prod"]
}'
```

## Delete
```sh
kubectl delete namespace application-fake-application-123
rm -rf /tmp/dolittle-local-dev/Source/V3/platform-api/453e04a7-4f9d-42f2-b36c-d51fa2c83fa3/fake-application-123
```