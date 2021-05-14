package microservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

func (s *service) handleBusinessMomentsAdaptor(w http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {

	var ms HttpInputBusinessMomentAdaptorInfo
	err := json.Unmarshal(inputBytes, &ms)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	tenant := k8s.Tenant{
		ID:   applicationInfo.Tenant.ID,
		Name: applicationInfo.Tenant.Name,
	}

	// TODO remove when happy with things
	if tenant.ID != "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3" {
		utils.RespondWithError(w, http.StatusBadRequest, "Currently locked down to tenant 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3")
		return
	}
	// TODO check tenantID with tenantID in the header

	application := k8s.Application{
		ID:   applicationInfo.ID,
		Name: applicationInfo.Name,
	}

	// TODO get from list in the cluster
	domainPrefix := "freshteapot-taco"
	ingress := k8s.Ingress{
		Host:       fmt.Sprintf("%s.dolittle.cloud", domainPrefix),
		SecretName: fmt.Sprintf("%s-certificate", domainPrefix),
	}

	if tenant.ID != ms.Dolittle.TenantID {
		utils.RespondWithError(w, http.StatusBadRequest, "tenant id in the system doe not match the one in the input")
		return
	}

	if application.ID != ms.Dolittle.ApplicationID {
		utils.RespondWithError(w, http.StatusInternalServerError, "Currently locked down to applicaiton 11b6cf47-5d9f-438f-8116-0d9828654657")
		return
	}

	//utils.RespondWithError(w, http.StatusBadRequest, "Before Create")
	//return
	namespace := fmt.Sprintf("application-%s", application.ID)
	err = s.businessMomentsAdaptorRepo.Create(namespace, tenant, application, ingress, ms)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO this could be an event
	// TODO this should be decoupled
	storageBytes, _ := json.Marshal(ms)
	err = s.gitRepo.Write(
		ms.Dolittle.TenantID,
		ms.Dolittle.ApplicationID,
		ms.Environment,
		ms.Dolittle.MicroserviceID,
		storageBytes,
	)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, ms)
}
