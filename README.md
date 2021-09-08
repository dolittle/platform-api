<h1 align="center"><img src="https://raw.githubusercontent.com/dolittle/Documentation/master/Source/static/images/dolittle_negativ_horisontal_RGB_black.svg" alt="Dolittle"></h1>

<h4 align="center">
    <a href="https://dolittle.io">Documentation</a> |
    <a href="https://dolittle.io/docs/tutorials/getting_started/">Tutorial</a> |
    <a href="https://github.com/dolittle/DotNet.SDK">C# SDK</a> |
    <a href="https://github.com/dolittle/JavaScript.SDK">JavaScript SDK</a>
</h4>

---

<p align="center">
    <a href="https://hub.docker.com/r/dolittle/platform-api"><img src="https://img.shields.io/docker/v/dolittle/platform-api?label=dolittle%2Fplatform-api&logo=docker&sort=semver" alt="Latest Docker image"></a>
    <a href="https://github.com/dolittle/platform-api/actions/workflows/ci.yml"><img src="https://github.com/dolittle/platform-api/actions/workflows/ci.yml/badge.svg" alt="Build status"></a>
</p>

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
