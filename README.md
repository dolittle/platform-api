# platform-api
What is a platform without an api? :)



## Update image
# Manual Update Image

```sh
docker build -f ./Dockerfile -t dolittle/platform-api:dev-x .
```

```sh
docker tag dolittle/platform-api:dev-x 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/platform-api:dev-x
```

```sh
docker push 508c17455f2a4b4cb7a52fbb1484346d.azurecr.io/dolittle/platform/platform-api:dev-x
```
