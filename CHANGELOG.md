# [4.8.1] - 2022-5-10 [PR: #125](https://github.com/dolittle/platform-api/pull/125)
## Summary

Resolved remaining comments of [PR 121](https://github.com/dolittle/platform-api/pull/121)

### Changed

- use logcontext
- log as info for validation messages
- refactoring based on [comments](https://github.com/dolittle/platform-api/pull/121)


# [4.8.0] - 2022-5-6 [PR: #121](https://github.com/dolittle/platform-api/pull/121)
## Summary

- Config Files support

### Added

- Get config files names list for microservice
  - /live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/config-files/list
    - GET
- Delete config file for microservice
  - /live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/config-files
    - DELETE
- Add entry to config file configmap for microservice
  - /live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/config-files
    - PUT
-  validation

### Changed

- More descriptive error logging in some of the environment variables repo and service


# [4.7.0] - 2022-5-5 [PR: #116](https://github.com/dolittle/platform-api/pull/116)
# Note
- This does not work in local development due to the container registry, we will need to document how to cheat the system

# Todo
- [ ] Write code to talk to azure
- [ ] Maybe a test for the http layer
- [ ] Think about creating storage interface for getting Tenant to reduce the testing


## Summary

Query the platform-api for images and tags in the container registry offered for each customer

### Added

- Get images
- Get tags based on image name


# [4.6.2] - 2022-5-4 [PR: #117](https://github.com/dolittle/platform-api/pull/117)
## Summary

Fix error when trying to create a base microservice without the M3Connector and the environment variable set.


# [4.6.1] - 2022-4-29 [PR: #114](https://github.com/dolittle/platform-api/pull/114)
## Summary

- Remove unused flag
- Fix command description


# [4.6.0] - 2022-4-28 [PR: #113](https://github.com/dolittle/platform-api/pull/113)
## Summary

- Listen for changes on where we will write connection details for m3connector and update the application
- Within application a new connection object exists in environment, this is used by the frontend today to provide a simple user experience
- When creating a microservice we will now add the m3connector volumes to the deployment


# [4.5.0] - 2022-4-25 [PR: #112](https://github.com/dolittle/platform-api/pull/112)
## Summary

Adds a new `HttpInputSimpleCommand` struct to the `Extra` field on the request for microservice creation. If the fields aren't specified the `command` and `args` fields will be emitted from the created deployment as the properties have the `omitempty` tag.

## Reference
- https://github.com/dolittle-platform/Operations/pull/209
- https://app.asana.com/0/1202121266838773/1202159723413336


# [4.4.0] - 2022-4-25 [PR: #106](https://github.com/dolittle/platform-api/pull/106)
## Summary

- List emails in an application in Azure
- Add email to an application in Azure
- Remove email from an application in Azure
- New endpoint to get customer details for one customer


# [4.3.1] - 2022-4-20 [PR: #111](https://github.com/dolittle/platform-api/pull/111)
## Summary

Creating a new application and it's environments now also ensures that the fileshares exist in Azure. Also fixes that the fileshares are correctly named in the `<application>-<environment>-backup` format.


# [4.3.0] - 2022-4-7 [PR: #110](https://github.com/dolittle/platform-api/pull/110)
## Summary

- When creating an application, it is possible to define customTenants per environment
- Creating applications uses ./modules/dolittle-application-with-resources.
- Fixes the workflow
- Retrigger release for https://github.com/dolittle/platform-api/pull/109


# [4.2.0] - 2022-4-6 [PR: #107](https://github.com/dolittle/platform-api/pull/107)
## Summary

When creating a new microservice that uses the [V8.0.0](https://github.com/dolittle/Runtime/releases/tag/v8.0.0) Runtime, we default to setting the `DOLITTLE__RUNTIME__EVENTSTORE__BACKWARDSCOMPATIBILITY__VERSION` env variable to `V7`. This is set through the `appsettings.json`, where the ASP.NET configuration system will pick it up just like it would be set in the env variable.
We default to `V7`, because to `V6` should only be used when upgrading a V6 Runtime to V8 and we don't currently support upgrading of the Runtime in the platform.

## Reference
- https://app.asana.com/0/1202023614957161/1202062723906543
- https://github.com/dolittle/Runtime/releases/tag/v8.0.0
- https://github.com/dolittle/Studio/pull/170


# [4.1.0] - 2022-4-1 [PR: #105](https://github.com/dolittle/platform-api/pull/105)
## Summary

Able to define the port that the head container will listen on, applied to service and deployment


# [4.0.0] - 2022-3-31 [PR: #102](https://github.com/dolittle/platform-api/pull/102)
## Summary

It's now possible to create _private microservices_, aka microservices that don't have an ingress and a domain. These microservices can only be accessed from inside the platform through their k8s Services DNS name.

To create a private microservice, the payload to send to the `/microservice` endpoint has a new `isPublic` property. Here's an example of such a request:
```sh
curl -X POST localhost:8081/microservice \
-H 'x-shared-secret: FAKE' \
-H 'Tenant-ID: 9049c278-7f6b-4fb9-a36e-c5223ca42fb5' \
-H 'User-ID: local-dev' \
-H "Content-Type: application/json" \
-d '
{
  "dolittle": {
    "applicationId": "286ee5a3-a41a-44b2-8f4f-bef9e61fe30b",
    "customerId": "e38dc77b-cf71-4e84-a361-28ab0de36ca7",
    "microserviceId": "1938f34a-ad5e-4de1-8196-bd3aab6f4954"
  },
  "name": "PrivateTest",
  "kind": "simple",
  "environment": "Dev",
  "extra": {
    "headImage": "nginxdemos/hello:latest",
    "runtimeImage": "dolittle/runtime:7.7.1",
    "ingress": {
      "path": "/",
      "pathType": "Prefix"
    },
    "isPublic": false
  }
}'
``` 

This is a breaking change as the default behaviour of creating a microservice has changed. If you don't supply the `isPublic` property in your request, platform-api will create a private microservice by default. 

The information saved to the git repo for the microservices `ms_*.json` file also includes the `isPublic` field, as the data is based of the changed `HttpInputSimpleInfo` struct. This means that requests to fetch a microservices data might also include the `isPublic` field in them, depending if they were created by platform-api before or after this change (unless a migration script is made). If microservice doesn't include an `isPublic` field it should be treated as if it were public with an ingress and a domain.

## Reference
- https://github.com/dolittle-platform/Operations/pull/192
- For my testing steps see https://github.com/dolittle-platform/Operations/pull/192#issuecomment-1077587867
- https://app.asana.com/0/1201955720774352/1201993650094129/f


# [3.5.1] - 2022-3-30 [PR: #66](https://github.com/dolittle/platform-api/pull/66)
## Summary
- Able to list-emails in the user system
- Able to link a customer to a user by email
- Able to link a customer to a user by user id
- Able to remove link from customer to user


# [3.5.0] - 2022-3-15 [PR: #98](https://github.com/dolittle/platform-api/pull/98)
## Summary

Fix bug when running "api update-repo".


# [3.4.0] - 2022-3-15 [PR: #95](https://github.com/dolittle/platform-api/pull/95)
## Summary

Adds new endpoints:
- GET `/studio/customer/{customerID}` for fetching a single customer's `studio.json` file. The configuration in the response is camelCased so that it's nicer to consume in JavaScript/TypeScript.
- POST `/studio/customer/{customerID}` for updating a single customer's `studio.json` file. The request should be camelCased.

Example request payload for disabling all environments:
```json
{"buildOverwrite":true,"disabledEnvironments":["*"],"canCreateApplication":true}
```

Curl example:
```sh
echo -n '{"buildOverwrite":true,"disabledEnvironments":[],"canCreateApplication":true}' \ 
| curl -X POST localhost:8081/studio/customer/CUSTOMER_ID \
 -H "x-shared-secret: FAKE" \
 -H "user-id: DEVELOPMENT_USER_ID" \
 -H "tenant-id: DEVELOPMENT_CUSTOMER_ID" \
 -H "Content-Type: application/json" \
 --data-binary @-
```

These endpoints enable us to modify the `studio.json`, so that we can enable or disable the creation of microservices per customer more easily.

## Reference
- [Related Studio PR](https://github.com/dolittle/Studio/pull/159)
- [Asana](https://app.asana.com/0/1201885260673030/1201894099742054)


# [3.3.0] - 2022-3-3 [PR: #97](https://github.com/dolittle/platform-api/pull/97)
## Summary

Bug creating a microservice, is rejecting if the ingressPath is used in a different environment


# [3.2.0] - 2022-3-3 [PR: #94](https://github.com/dolittle/platform-api/pull/94)
## Summary

- Adding dolittle-config configmap specific for dolittle/runtime:6.1.0.


# [3.1.0] - 2022-3-3 [PR: #96](https://github.com/dolittle/platform-api/pull/96)
## Summary

- job to build platform-api
- job to build platform-operations


# [3.0.0] - 2022-2-25 [PR: #90](https://github.com/dolittle/platform-api/pull/90)
## Summary

Changed the CLI commands:
- New root command `platform template`.
	- `delete` for creating templates that delete things platform wide. Replaces old `platform tools studio delete-*` commands:
		- `application`
		- `customer`
- Refactored `platform tools studio` command:
  - `upsert` as all of these commands are about upsert the existing `*.json` files.
    - `terraform`
    - `studio`
    - `application`
  - `create` for creating new resources
    - `service-account`
  - `get` for getting Studio specific resources
    - `customers`
- Refactor `platform tools job` command:
	- `status` get's the job status
	- `template` for creating the pre-filled k8s Jobs
		- `customer`
		- `application`
- Refactor `platform tools terraform` command:
	- `template` for creating the pre-filled Terraform HCL
		- `customer`
		- `application`
- Moved `platform tools automate get-microservice-metadata` command to `platform tools explore microservices`

## Ref
- [Brainstorming Miro board](https://miro.com/app/board/uXjVOKmmIm4=/)
- [Asana ticket](https://app.asana.com/0/home/1142702759924347/1201235587155608)


# [2.12.0] - 2022-2-22 [PR: #92](https://github.com/dolittle/platform-api/pull/92)
## Summary

- Moving mocks into own package and adding makefile command to make it easier to update interfaces `make build-mocks`.
- Adding codeclimate yaml


# [2.11.0] - 2022-2-22 [PR: #91](https://github.com/dolittle/platform-api/pull/91)
## Summary

Rebuilding application state, take status + welcomeMicroservice from storage if exists.


# [2.10.3] - 2022-2-22 [PR: #89](https://github.com/dolittle/platform-api/pull/89)
## Summary

Reducing the mixture of tenantID and customerID


# [2.10.2] - 2022-2-22 [PR: #88](https://github.com/dolittle/platform-api/pull/88)
## Summary

Rename from GetTenantDirectory to GetCustomerDirectory


# [2.10.1] - 2022-2-18 [PR: #87](https://github.com/dolittle/platform-api/pull/87)
## Summary
Fixing bug in delete microservice endpoint


# [2.10.0] - 2022-2-17 [PR: #86](https://github.com/dolittle/platform-api/pull/86)
## Summary

- Able to add user to "studio admin"
- Able to remove user from "studio admin"
- Able to list users who have "studio admin"


# [2.9.0] - 2022-2-16 [PR: #56](https://github.com/dolittle/platform-api/pull/56)
## Summary

Exposing the raw subjectRulesReviewStatus for a user, to be manipulated via /application/{applicationID}/personalised-application-info.


# [2.8.0] - 2022-2-11 [PR: #59](https://github.com/dolittle/platform-api/pull/59)
# Summary
- Able to create a customer via the command line
- Able to create a customer via the rest api
- Able to create an application via the command line
- Able to create an application via the rest api
- Able to create a microservice
- Developer rbac out the box, supports querying own namespace
- Creating an application will make a namespace
- Add Acr secret to access the container registry
- Application Storage
- Application Role
- Application RoleBindings
- Environment Networkpolicy
- Environment tenants configmap
- Mongo (Service, Stateful, Cronjob)
- For local dev we create rolebindings for the local-dev user (Linked to Studio and k3d)


# [2.7.0] - 2022-2-2 [PR: #83](https://github.com/dolittle/platform-api/pull/83)
## Summary

- Reducing the amount of information used for the personalised-info, to that which is currently used in Studio.
- `GetStudioInfo` function to reduce the boiler plate of getting requried info per request.


# [2.6.0] - 2022-2-2 [PR: #82](https://github.com/dolittle/platform-api/pull/82)
## Summary

Including the environment name in the lookup for microservice name


# [2.5.3] - 2022-1-28 [PR: #81](https://github.com/dolittle/platform-api/pull/81)
## Summary

Reducing the amount of place we copy the same code to get the client for kubernetes.
Now we can use `k8sClient, k8sConfig := platformK8s.InitKubernetesClient()`


# [2.5.2] - 2022-1-28 [PR: #80](https://github.com/dolittle/platform-api/pull/80)
## Summary

Fixes the `add-platform-config` command to skip deployments that don't have a Runtime container.

Also skips if the microservice doesn't have a *-dolittle configmap.

Also checks first if the deployment the volumeMount would be added to has a corresponding volume specified, otherwise the deployment would fail to start. In testing with dry-run I didn't see that logmessage pop up but it should still be handled just to be sure.


# [2.5.1] - 2022-1-27 [PR: #79](https://github.com/dolittle/platform-api/pull/79)
## Summary

Changes the `GetDeployment` function to require `applicationID` & `environment` instead of just a `namespace`. This is because some microservices reuse the same microserviceID accross environment. This makes it so that you can reliably always get the correct deployment from the correct environment.


# [2.5.0] - 2022-1-27 [PR: #73](https://github.com/dolittle/platform-api/pull/73)
## Summary

- Adds new command `pull-dolittle-config` which pulls all of the `*-dolittle` configmaps from the cluster and writes them into the `Source/V3/Kubernetes/Customers/<customer>/<application>/<environment>/<microservice>/microservice-configmap-dolittle.yml` files. Related PR https://github.com/dolittle-platform/Operations/pull/150
Usage:
```sh
go run main.go tools automate pull-dolittle-config <repo-root>
```

- Adds new command `add-platform-config` which updates the `*-dolittle` configmaps in the cluster with the new `platform.json` resource and also updates the microservices deployments Runtime container to have a volumeMount for the `platform.json`
Usage:
```sh
	go run main.go tools automate add-platform-config \
	--application-id="11b6cf47-5d9f-438f-8116-0d9828654657" \
	--environment="Dev" \
	--microservice-id="ec6a1a81-ed83-bb42-b82b-5e8bedc3cbc6" \
	--dry-run=true
```

- Adds new command `get-microservices-metadata` which outputs a JSON array of all of the metadata of all of the clusters microservices

- Adds new command `pull-microservice-deployment` which pulls all microservices deployments from the cluster and writes them into `Source/V3/Kubernetes/Customers/<customer>/<application>/<environment>/<microservice>/microservice-deployment.yml` files. Related PR https://github.com/dolittle-platform/Operations/pull/155
Usage:
```sh
go run main.go tools automate pull-microservice-deployment <repo-root>
```

- Adds new command `import-dolittle-configmaps` which ca n be used to bootstrap your local cluster. From PR https://github.com/dolittle/platform-api/pull/74

- When pulling the deployments or configmaps and converting the given resource to a `configK8s.Microservice` struct, if the given resource doesn't have a `dolittle.io/microservice-kind` specified it will default to the new `MicroserviceKindUnknown`. This is only internal for now and it's not set to the k8s resources annotations.
- Adds new package `automate` to help with the `platform tools automate` commands


# [2.4.1] - 2022-1-26 [PR: #77](https://github.com/dolittle/platform-api/pull/77)
## Summary

Fixes the `getServiceAccountCredentials()` method to return the whole k8s Secret resource. This is because Azure DevOps requires the whole Secret resource instead of just the data part when setting up a kubernetes connection.

The Studio button that calls this endpoint:
![image](https://user-images.githubusercontent.com/10163775/150506082-c26b1701-4f4f-481b-8af4-8021e3a52d66.png)


# [2.4.0] - 2022-1-25 [PR: #78](https://github.com/dolittle/platform-api/pull/78)
## Summary

Include platform_environment in the sub-folder for the location of the studio files within git storage.

- Building json files now includes platform-environment
- Removed unused shouldCommit
- Removed old code examples
- Fixed a wrong variable
- Moved all to be local for its specific function
- Introduced disable-environments flag for studio.json creation
- Make go-staticcheck happy plus moving studio commands into own sub-package


### Added
- `--platform-environment` to tools studio {build-studio-info, build-terraform-info, build-applicationinfo-info}
- The server via PLATFORM_ENVIRONMENT aware of where to get the data for studio from.

### Changed
- moved `microservice build-terraform-info` to `tools studio build-terraform-info`
- moved `microservice build-studio-info` to `tools studio build-studio-info`
- moved `microservice build-application-info` to `tools studio build-application-info`
- `tools studio build-studio-info` has the ability to disable-environments via the `--disable-environments` flag


# [2.3.3] - 2022-1-11 [PR: #71](https://github.com/dolittle/platform-api/pull/71)
## Summary
- Disable creation of microservice, whilst we fix a few bugs


# [2.3.2] - 2021-12-17 [PR: #70](https://github.com/dolittle/platform-api/pull/70)
## Summary

- api to get the current environment variables in a Studio friendly manner
- api to update environment variables from the Studio


# [2.3.1] - 2021-12-15 [PR: #69](https://github.com/dolittle/platform-api/pull/69)
## Summary

- Api to restart microservice


# [2.3.0] - 2021-12-15 [PR: #68](https://github.com/dolittle/platform-api/pull/68)
## Summary

- Building the application info to include ModuleName and ApplicationID.
- Terraform create customer now includes the module_name.


# [2.2.0] - 2021-12-13 [PR: #67](https://github.com/dolittle/platform-api/pull/67)
## Summary

Added a few more attributes to the terraformCustomer based on information from the Operations repo.


# [2.1.3] - 2021-12-1 [PR: #65](https://github.com/dolittle/platform-api/pull/65)
## Summary

- Rename HttpResponseApplications2 to HttpResponseApplication
- Expose the Ingress host + path per customer Tenant ID
- Expose Unique Ingress Paths (microservice paths)


# [2.1.2] - 2021-11-29 [PR: #62](https://github.com/dolittle/platform-api/pull/62)
## Summary

Sharing the git InitGit command

Linked to https://app.asana.com/0/1201338547604691/1201417201813836/f


# [2.1.1] - 2021-11-24 [PR: #63](https://github.com/dolittle/platform-api/pull/63)
## Summary

Add a containerPort called `runtime-metrics` on 9700 when creating a Runtime container. This is so that we can scrape the Runtime metrics from it.

Also adds the `metrics.json` configuration file to the Runtime's configurations. Currently it's not read, aka changes to it won't reflect on the `runtime-metrics` containerPort (just like how `endpoint.json` also doesn't affect the actual opened ports of the container).


# [2.1.0] - 2021-11-23 [PR: #61](https://github.com/dolittle/platform-api/pull/61)
## Summary

Add's the ability to create a base microservice without a Runtime container by defining the wanted Runtime image as `"none"` in the request.

Also adds memory limits to newly created Runtime containers in a microservice. The request is for 250MB and the limit is 1GB.

Related to https://github.com/dolittle/Studio/pull/136


# [2.0.0] - 2021-11-11 [PR: #60](https://github.com/dolittle/platform-api/pull/60)
## Summary

Changes the `create-service-account` command to create a new `devops` rolebinding instead of adding to the existing `developer` rolebinding.

This is so that we don't have to worry about modifying the `developer` rolebinding in the future.


# [1.4.0] - 2021-11-11 [PR: #58](https://github.com/dolittle/platform-api/pull/58)
## Summary

Adds the new  `create-service-account` command, which creates a  service account called `devops` and adds it to the `developer` role binding


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


