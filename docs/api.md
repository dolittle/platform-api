
# Live from the cluster

# Get applications by tenant
```sh
curl -XGET localhost:8080/live/tenant/453e04a7-4f9d-42f2-b36c-d51fa2c83fa3/applications
```

# Get microservices by application
```sh
curl -XGET localhost:8080/live/application/fe7736bb-57fc-4166-bb91-6954f4dd4eb7/microservices | jq
```