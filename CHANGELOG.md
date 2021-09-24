# [0.5.2] - 2021-9-24 [PR: #33](https://github.com/dolittle/platform-api/pull/33)
## Summary

While a pod I starting up, the StartTime is not set (nil in Go) which caused the "View" screen just after creating e.g. a POapi to fail. This fix just returns "N/A" instead of an actual date when StartTime is nil. The contract was just a plain string anyways, so Studio will accept it.

#### How to test
- Start with a clean cluster (or without raw-data-log and purchase-order)
- Click Microservices
- Click Create new microservice
- Click Create within the Purchase order api card
- Select Infor graphic
- Enter name
- Click next
- Enter username
- Enter password
- Click Next
- At this point, we send the request to the backend and redirect the frontend.
- Observe platform-api __does not__ crash

### Fixed

- Handling of Pod.StartTime nil-values while the pod is still starting up.


# [0.5.1] - 2021-9-24 [PR: #34](https://github.com/dolittle/platform-api/pull/34)
## Summary

Fixes some typos in the HTTP responses in PurchaseOrderAPI handler

### Fixed

- Typos in POAPI handler HTTP responses


# [0.5.0] - 2021-9-22 [PR: #31](https://github.com/dolittle/platform-api/pull/31)
## Summary

Summary of the PR here. The GitHub release description is created from this comment so keep it nice and descriptive.

Remember to remove sections that you don't need or use.

### Added

- Describe the added features

### Changed

- Describe the outwards facing code change

### Fixed

- Describe the fix and the bug

### Removed

- Describe what was removed and why

### Security

- Describe the security issue and the fix

### Deprecated

- Describe the part of the code being deprecated and why


# [0.4.3] - 2021-9-20 [PR: #30](https://github.com/dolittle/platform-api/pull/30)
## Summary

Adds checks to ensure that there doesn't already exist a PurchaseOrderAPI in an environment when creating. The previous logic checked for one with the same name and id, the added code checks for any name and id.

The code might be a bit excessive, as it now first checks the Git storage, and then K8s afterwards. We could get away with one of them, but I'm not sure which one to truly consider the source of truth until we have everything covered by automation. Also might not be clear, but the old check is still useful because it checks for any microservice with the same name (which would cause problems) - even though the naming of the function kind of indicates that it looks for a PurchaseOrderAPI microservice.

I've tested it locally, and also added more specs for the `Exists` method and the new `EnvironmentHasPurchaseOrderAPI` method.

### Added

- Disallow the creation of multiple PurchaseOrderAPIs in the backend.

### Changed

- The structure of the specs for the purchase order api repo


# [0.4.2] - 2021-9-17 [PR: #28](https://github.com/dolittle/platform-api/pull/28)
## Summary

Add handling of an error

### Fixed

- An error that was not handled
- Moved saving of purchase order api to git repo right after the microservice is created


# [0.4.1] - 2021-9-17 [PR: #25](https://github.com/dolittle/platform-api/pull/25)
## Summary

When getting applications from git repo start walking the directory from the customer folder, instead of the root of the repo.

### Changed

- When getting applications from git repo start searching from the customer folder
- Use the logger


# [0.4.0] - 2021-9-17 [PR: #24](https://github.com/dolittle/platform-api/pull/24)
## Summary

Checks whether a purchase order api microservice already exists by comparing the name of the deployment If it already exists it will return with an error and write a response with StatusInternalServerError

### Added

- Check for whether purchase order api exists when creating it

### Changed

- Changed some parameters and variables from *kubernetes.ClientSet to kubernetes.Interface in order to write specs
- Temporarily removed the dependency to rawDataLogIngestorRepo from purchaseorderapi repo in order to make it easier to write specs


# [0.3.0] - 2021-9-16 [PR: #26](https://github.com/dolittle/platform-api/pull/26)
## Summary

When creating a PurchaseOrderAPI microservice, if the environment doesn't already have a RawDataLog microservice and Stan & Nats it will create them.

### Added

- Creates the missing RawDataLog deployment and Nats & Stan statefulsets when creating a PurchaseOrderAPI microservice if they didn't already exist
- Save the RawDataLog microservice to the git repo
- Add logging for the creation of RawDataLog

### Changed

- Describe the outwards facing code change

### Fixed

- Fixes the `NATS_CLUSTER_URL` env to the hardcoded `<env>-nats.namespace...` format
- Also makes the RawDataLog handler (the specific one) to check for the RawDataLogs existence beforehand


# [0.2.2] - 2021-9-16 [PR: #20](https://github.com/dolittle/platform-api/pull/20)
## Summary

I've changed the setup of NATS+STAN resources in the RawDataLogRepo:
1. The name now includes the environment (to allow more than one)
2. Removed the pod-anti-affinity settings on the STAN statefulset so that we can have more than 3 in total
3. Added the Dolittle labels to the statefulsets (and pod templates) so that they are covered by the network policies
4. Changed the setup so they are created with the non-dynamic Kubernetes client (to make it easier to test - the dynamic client requires a connection to a cluster to work).

The tests are quite big, but I'm not verse with Ginkgo yet, so improvements on the structure are welcome :)

I tried this out locally, and it does not work with the k3d cluster (the nats pod doesn't allow any connections). But with the docker cluster it works.

#### Asana tasks
- https://app.asana.com/0/1200747172416688/1200762338877903/f
- https://app.asana.com/0/1200879032451150/1200931738686016/f

### Changed

- The creation of NATS+STAN resources now includes environment in the name, has labels to work with existing networkpolicies
- Change creation of RawDataLog for NATS+STAN resources to use the non-dynamic (hardcoded) Kubernetes client.
- Changed `*kubernetes.ClientSet` to `kubernetes.Interface` everywhere to enable the use of the Fake client.

### Fixed

- The wrong tenant id was given when creating config map for rawdatalogingestor microservice


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


