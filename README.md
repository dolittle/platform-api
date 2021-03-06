<h1 align="center"><img src="https://raw.githubusercontent.com/dolittle/Documentation/master/Source/static/images/dolittle_negativ_horisontal_RGB_black.svg" alt="Dolittle"></h1>

<h4 align="center">
    <a href="https://dolittle.io">Documentation</a> |
    <a href="https://dolittle.io/docs/tutorials/getting_started/">Tutorial</a> |
    <a href="https://github.com/dolittle/DotNet.SDK">C# SDK</a> |
    <a href="https://github.com/dolittle/JavaScript.SDK">JavaScript SDK</a>
</h4>

---

<p align="center">
    <a href="https://hub.docker.com/r/dolittle/platform-api"><img src="https://img.shields.io/docker/v/dolittle/platform-api/latest?label=dolittle%2Fplatform-api&logo=docker" alt="Latest Docker image"></a>
    <a href="https://github.com/dolittle/platform-api/actions/workflows/ci.yml"><img src="https://github.com/dolittle/platform-api/actions/workflows/ci.yml/badge.svg" alt="Build status"></a>
    <a href="https://codeclimate.com/github/dolittle/platform-api/maintainability"><img src="https://api.codeclimate.com/v1/badges/ca8fcdc442c82f0e0982/maintainability" /></a>
    <a href="https://codeclimate.com/github/dolittle/platform-api/test_coverage"><img src="https://api.codeclimate.com/v1/badges/ca8fcdc442c82f0e0982/test_coverage" /></a>
</p>

# platform-api
What is a platform without an api? :)

# Overview

`platform-api` is a project to automate the [Dolittle Platform](https://dolittle.io/docs/platform/).

It is built from 3 main CLI tools:

- `api` is a CLI tool that builds JSON files from our cluster. It's also a server, that handles k8s resources (get, create, etc). It lives in `cmd/api`.

- `rawdatalog` is a:
    - code entry point into the raw-data-log
    - sharing code
    - lives in `cmd/rawdatalog`

- `tools` is a collection of different commands to help us manage the platform. It lives in `cmd/tools`.

# Setup

Check the tutorial from [Studio](https://github.com/dolittle/Studio/blob/main/Documentation/k3d.md) on how to set it up with a local cluster.

# Testing

Run the tests:
```sh
go test -v ./...
```

For creating/updating the mocks you'll need [mockery](https://github.com/vektra/mockery).
```sh
make build-mocks
```
