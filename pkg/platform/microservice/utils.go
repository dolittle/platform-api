package microservice

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

// Reads a the microservice info from input by unmarshaling the json into the first argument.
func readMicroservice(microservice *platform.HttpMicroserviceBase, input []byte, applicationInfo platform.Application) (microserviceK8sInfo, error) {
	err := json.Unmarshal(input, &microservice)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return microserviceK8sInfo{}, err
	}

	tenant := k8s.Tenant{
		ID:   applicationInfo.Tenant.ID,
		Name: applicationInfo.Tenant.Name,
	}

	application := k8s.Application{
		ID:   applicationInfo.ID,
		Name: applicationInfo.Name,
	}
	if tenant.ID != microservice.Dolittle.TenantID {
		utils.RespondWithError(w, http.StatusBadRequest, "tenant id in the system doe not match the one in the input")
		return
	}

	if application.ID != microservice.Dolittle.ApplicationID {
		utils.RespondWithError(w, http.StatusInternalServerError, "Currently locked down to applicaiton 11b6cf47-5d9f-438f-8116-0d9828654657")
		return
	}

	namespace := fmt.Sprintf("application-%s", application.ID)
}
