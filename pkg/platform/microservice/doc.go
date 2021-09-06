package microservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

type service struct {
	simpleRepo                 simpleRepo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	rawDataLogIngestorRepo     rawdatalog.RawDataLogIngestorRepo
	purchaseOrderAPIRepo       purchaseOrderAPIRepo
	k8sDolittleRepo            platform.K8sRepo
	gitRepo                    storage.Repo
}

type microserviceK8sInfo struct {
	tenant      k8s.Tenant
	application k8s.Application
	namespace   string
}

// Reads a the microservice info from input by unmarshaling the json into the first argument.
// Writes an error response if the reading fails for any reason, setting the returning success false.
func readMicroservice(microservice platform.Microservice, input []byte, applicationInfo platform.Application, responseWriter http.ResponseWriter) (info microserviceK8sInfo, success bool) {
	info = microserviceK8sInfo{}

	err := json.Unmarshal(input, &microservice)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return info, false
	}

	info.tenant = k8s.Tenant{
		ID:   applicationInfo.Tenant.ID,
		Name: applicationInfo.Tenant.Name,
	}

	info.application = k8s.Application{
		ID:   applicationInfo.ID,
		Name: applicationInfo.Name,
	}
	if info.tenant.ID != microservice.GetBase().Dolittle.TenantID {
		utils.RespondWithError(responseWriter, http.StatusBadRequest, "tenant id in the system doe not match the one in the input")
		return info, false
	}

	if info.application.ID != microservice.GetBase().Dolittle.ApplicationID {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, "Currently locked down to applicaiton 11b6cf47-5d9f-438f-8116-0d9828654657")
		return info, false
	}

	info.namespace = fmt.Sprintf("application-%s", info.application.ID)
	return info, true
}
