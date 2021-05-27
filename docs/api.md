

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
## Post

type BusinessMoment struct {
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
	Filter    string `json:"filter"`    // filter https://github.com/dolittle-entropy/no-code-experiment/blob/demo-prep/src/editor.svelte
	Mapper    string `json:"mapper"`    // mapper https://github.com/dolittle-entropy/no-code-experiment/blob/demo-prep/src/editor.svelte
	Transform string `json:"transform"` // businessMomentsMapper https://github.com/dolittle-entropy/no-code-experiment/blob/demo-prep/src/editor.svelte
}

type HttpInputBusinessMoment struct {
	ApplicationID  string         `json:"application_id"`
	Environment    string         `json:"environment"`
	MicroserviceID string         `json:"microservice_id"`
	Moment         BusinessMoment `json:"moment"`
}


```sh
curl -XPOST localhost:8080/businessmoment -d '
{
  "applicationId": "11b6cf47-5d9f-438f-8116-0d9828654657",
  "environment": "Dev",
  "microserviceId": "934789f9-c1c0-2f4e-acb4-49a3b59b7e50",
  "moment": {
    "name": "I am name",
    "uuid": "fake-uuid-123",
    "filter": "filter func",
    "mapper": "mapper func",
    "transform": "transform func"
  }
}'
```
