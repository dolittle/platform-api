# [1.3.5] - 2021-11-10 [PR: #57](https://github.com/dolittle/platform-api/pull/57)
## Summary

- Make the platform-api aware of the external host of the cluster for environment variable


# [1.3.4] - 2021-11-10 [PR: #55](https://github.com/dolittle/platform-api/pull/55)
## Summary

- Added a makefile to make it easier to test the project
- Included more code paths into the tests
- Show total test coverage via "make test".


# [1.3.3] - 2021-11-10 [PR: #54](https://github.com/dolittle/platform-api/pull/54)
## Summary

- Adding management port to endpoints.json


# [1.3.2] - 2021-11-9 [PR: #52](https://github.com/dolittle/platform-api/pull/52)
## Summary
- Specific endpoint for devops to not expose all service accounts.
- Lookup service account to find the secret name.


# [1.3.1] - 2021-11-8 [PR: #53](https://github.com/dolittle/platform-api/pull/53)
## Summary

Change from github.com/dolittle-entropy/platform-api to github.com/dolittle/platform-api.


# [1.3.0] - 2021-11-4 [PR: #50](https://github.com/dolittle/platform-api/pull/50)
## Summary

- Endpoint for service-account credentials
- Endpoint for container-registry credentials


# [1.2.0] - 2021-11-4 [PR: #51](https://github.com/dolittle/platform-api/pull/51)
## Summary

Querying the /applications data, will now return an empty array, not null.


# [1.1.2] - 2021-11-4 [PR: #49](https://github.com/dolittle/platform-api/pull/49)
## Summary

- Formatting server
- Bringing middleware.RestrictHandlerWithSharedSecret

Linked to #43


# [1.1.1] - 2021-11-2 [PR: #48](https://github.com/dolittle/platform-api/pull/48)
## Summary

Making build-terraform-info aware of the platform-environment


# [1.1.0] - 2021-10-26 [PR: #47](https://github.com/dolittle/platform-api/pull/47)
## Summary

- Stub terraform block for a customer.
- Renamed root level to "platform".


# [1.0.0] - 2021-10-22 [PR: #46](https://github.com/dolittle/platform-api/pull/46)
## Summary

Makes the platform-api ready to start supporting automation for everybody by adding new commands and changing some current behavior.

Related Studio doc PR https://github.com/dolittle/Studio/pull/126

When deployed both `auto-prod` and `auto-dev` have to be fixed (in their respective environments):

- `auto-dev` https://github.com/dolittle-platform/Operations/pull/106
- `auto-prod` https://github.com/dolittle-platform/Operations/pull/107

### Added

- Adds a new global `--git-dry-run` flag an `GIT_REPO_DRY_RUN` env variable that disables committing and pushing. The `GitStorage` respects this and won't add files to the index, commit or push changes when it's set to true.
- New `build-studio-info` command to build default `studio.json` files. You can pass it the `customerID`  to default a specific customers studio config, or use the `--all` flag to reset it for all customers in the cluster. 

### Changed

- All files/directories are now written into `Source/V3/platform-api/`. This is to prepare for the Great Merge with Operations.
- Automation is enabled by default for all customers, now you have to opt-out in `studio.json` `"disabled_environments"` property. This changes `studio.json` to look like this:
```json
{
  "build_overwrite": true,
  "disabled_environments": [
	// list of disabled applications and their environments
	"38185bb0-33ed-4bde-a0af-7ac736055dd7/dev"
	]
}
```
- `--kube-config` flag and `KUBECONFIG` env variable is global for all `microservice *` commands.

### Fixed

- Renamed `build-applications.go` to `build-application-info.go` to reflect the actual name of the command
- `application.json` will use its own `JSONApplication` & `JSONEnvironment` structs instead of the `HttpResponse*` structs.
- If `GIT_REPO_DIRECTORY_ONLY` env variable is set to true platform-api won't try to pull.

### Removed

- Removed the `automationEnabled` property from the `"environments"` property in `application.json` as it didn't actually reflect the applications automation status.


# [0.9.0] - 2021-10-18 [PR: #45](https://github.com/dolittle/platform-api/pull/45)
## Summary

Makes platform-api pull from the remote before starting to do any changes. The push will only happen if the local repo is already up to date, or if it can be fast-forwarded.


### Added

- More logging

### Changed

- Pull before writing any files

### Removed

- `CheckAutomationEnabledViaCustomer` and replaced calls to it with the more common `IsAutomationEnabled`


# [0.8.0] - 2021-10-12 [PR: #44](https://github.com/dolittle/platform-api/pull/44)
## Summary
- Explicitly set to lookup kubeconfig within the cluster
- If the metadata.annotations["dolittle.io/microservice-kind"] exists it will include it "kind" in the /api/live/application/{applicationId}/microservices endpoint


# [0.7.1] - 2021-10-11 [PR: #42](https://github.com/dolittle/platform-api/pull/42)
## Summary

Fix the `microservice server` command to default to users default kubeconfig location `~/.kube/config` if no kubeconfig was defined.

Also fix a bug with settings not being correctly printed out during startup

### Changed

- `microservice server` command now defaults to users default kubeconfig location `~/.kube/config`

### Fixed

- Print out the `tools.server` settings correctly during startup


# [0.7.0] - 2021-10-8 [PR: #41](https://github.com/dolittle/platform-api/pull/41)
## Summary

Adds a new API endpoint `/application/{applicationID}/environment/{environment}/purchaseorderapi/{microserviceID}/datastatus` for fetching the current status of PurchaseOrderAPI through it's own API (related PR https://github.com/dolittle/ERPIntegrations/pull/37). The return payload looks like this:

```json
{"status":"Running","lastReceivedPayload":"2021-10-05T15:52:18.795Z","error":""}
```

### Added

- New API endpoint `/application/{applicationID}/environment/{environment}/purchaseorderapi/{microserviceID}/datastatus` for fetching the status of a PurchaseOrderAPI Microservice.
- Workflow for building and pushing a Docker image on every pull request

### Fixed

- Added the missing `projections` and `embeddings` sections to the created `resources.json` configmap for PurchaseOrderAPI creation


# [0.6.2] - 2021-9-30 [PR: #36](https://github.com/dolittle/platform-api/pull/36)
## Summary

Adds a PersistentVolumeClaim to the STAN deployment with the name of `<env>-stan-storage` for the PVC. It uses the same [`managed-premium`](https://docs.microsoft.com/en-us/azure/aks/concepts-storage#storage-classes) storage class as we use for MongoDB in the platform. The STAN store method is [file](https://docs.nats.io/nats-streaming-server/configuring/persistence/file_store) and the name of the directory is `datastore` like in the STAN doc examples.

To use this locally you have to create a new storage class with the name `managed-premium`, create a persistent volume attached to it, and attach a directory to that persistent volume during k3d creation. This is documented in the Studio PR https://github.com/dolittle/Studio/pull/119

### Added

- Adds a PersistentVolumeClaim of 8GB to the STAN deployment

### Changed

- Made some logging around NATS/STAN creation more clear in RawDataLog


# [0.6.1] - 2021-9-30 [PR: #39](https://github.com/dolittle/platform-api/pull/39)
## Summary

Fixes a bug where if an error occurs during creation of a PurchaseOrderAPI microservice, both the error object and the microservice object was written to the response. Resulting in invalid JSON in the response body.

### Fixed

- Don't append the microservice object to the response body if an error occurs during creation of a PurchaseOrderAPI


# [0.6.0] - 2021-9-29 [PR: #37](https://github.com/dolittle/platform-api/pull/37)
## Summary

This PR adds the functionality to change the webhooks configuration for purchase order api and raw data log, both when creating/deleting purchase order api and updating the username/password on an already existing purchase order api (and by extension raw data log).

### Added

- PUT /microservice route that uses the Update request handler on the microservice.Service struct
- Update method in microservice.Service that currently only can update the webhooks of a purchase order api. In the future we should consider how to extend this method to other microservice kinds, this implementation was just done to achieve to shortest path to the immediate goal.
- UpdateWebhooks method on purchaseorderapi.Handler that updates only the webhook configuration of the existing purchase order api and raw data log. It will return error if purchase order api does not exist, but it will create the raw data log if that doesn't exist (which shouldn't ever be the case)
- Update method on rawdatalog.Repo that only updates the config files config map of the raw data log ingestor repo. I'm not certain that this will ensure that the configuration will be refreshed for the deployment, though after talking with @jakhog we think that it will work this way.

### Changed

- Changed server endpoints to use the http.Method* constants instead of magic strings
- Added (and use) the logger to the microservice.Service struct
- Purchase Order API handler signature so that it does not take in the response writer, but instead returns a custom Error type that includes the error and the http status code. This is done because it should be the responsibility of the caller (microservice.Service) to respond with a http response.
- Change the behaviour of the purchaseorderapi.Handler Create method so that it ensures that a purchase order api does not exist and also ensures that a raw data log ingestor exists with the configured webhoooks from the purchase order api microservice information from the request. Meaning that it will create a new raw data log if it does not exist or update the existing one with the webhook configuration form the request.
- Exists method on rawdatalog.Repo now returns the microservice ID as well


# [0.5.4] - 2021-9-29 [PR: #38](https://github.com/dolittle/platform-api/pull/38)
## Summary

Fixes a minor issue where an array was not initialised to an empty array

### Changed

- Initialize pods list with empty array


# [0.5.3] - 2021-9-27 [PR: #35](https://github.com/dolittle/platform-api/pull/35)
## Summary

There was a bug in the raw data log ingestor repo implementation that converted the environment from the microservice from the input to lower case, causing the environment label on the raw data log deployment to be different from the other microservice kind deployments. 

However this fix has a side effect that I don't know the effects of, since we do not convert the environment to lower that results in the DOLITTLE_ENVIRONMENT environment variable on the envVariables config map being different (not lowercased). Is this important? If so we can just make the input there to lower case, that's not a big deal.
EDIT: As of writing this the code in question is on line 444


#### How to test
- Go through test steps in PR in Studio https://github.com/dolittle/Studio/pull/116

### Fixed

- Deployment of raw data log should now have the environment label in the proper casing


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


