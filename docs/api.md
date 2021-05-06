
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
