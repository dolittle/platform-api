# [0.2.0] - 2021-9-14 [PR: #19](https://github.com/dolittle/platform-api/pull/19)
## Summary

Changes the structure of the HttpInputEnvironment struct so that it knows about the tenants and has a mapping between tenant Id and ingress.

There are some assumptions in play around the ingresses that are added to the map. First it assumes that the ingresses that needs to be configured has one or more spec.TLS items. Secondly it will override the ingress config for a tenant. It also maybe clunkily extracts the tenant id from the ingress from the nginx.ingress.kubernetes.io/configuration-snippet annotation. 

### Changed

- Changed structure of HttpInputEnvironment to include a list of tenants and a mapping between tenant id and ingress host and domain prefix
- build-applications command to store applications with this new structure

# Reference
- https://app.asana.com/0/1200879032451150/1200931738686043/f


# [0.1.0] - 2021-9-9 [PR: #15](https://github.com/dolittle/platform-api/pull/15)
## Summary

Adds the functionality to create and delete purchase order api microservices

### Added

- Input types for PurchaseOrderAPI microservice for when doing a request to create a microservice
- MicroserviceKind for PurchaseOrderAPI
- Create purchase order api microservice
- Delete purchase order api microservice

### Fixed

- Some issues in regard to raw data log
- Bug related to MongoDB connection strings


# [0.0.1] - 2021-9-8 [PR: #18](https://github.com/dolittle/platform-api/pull/18)
## Summary

Fixes the filepath code to also work with Windows.

### Fixed

- Fixes the code to also understand Windows style path delimiters instead of just *nix paths when reading the repo.


