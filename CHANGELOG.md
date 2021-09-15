# [0.2.1] - 2021-9-15 [PR: #22](https://github.com/dolittle/platform-api/pull/22)
## Summary

When creating Purchase Order API microservice or RawDataLogIngestor microservice it is important to configure it to the correct tenant. The tenant being used before was just a static tenant, it needs to not be static in order for it to work with actual customers.

There are a few caveats at the moment:
- It will only use one tenant. For instance if we needed a Purchase Order API to commit events to multiple tenants that wouldn't work with the current solution
- The added configured tenant is not used for the other microservices yet. Only used for Purchase Order API and RawDataLogIngestor
- When we want to do this for normal microservices we need to change this to configure for all tenants, not just the first one 

### Added

- A method for getting tenant for environment on the application metadata stored in git
- Specs for method for getting tenant for environment on the application metadata stored in git 

### Changed

- The tenant used when creating Purchase Order API and RawDataLogIngestor
- A couple of places in the code 'tenant' parameter was renamed to 'customer' to make it more explicit that the tenant actually represents a Dolittle customer


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


