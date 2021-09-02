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

func (s *service) handleSimpleMicroservice(w http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place

	var ms platform.HttpInputSimpleInfo
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
	application := k8s.Application{
		ID:   applicationInfo.ID,
		Name: applicationInfo.Name,
	}

	// TODO replace this with something from the cluster or something from git
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

	// TODO I cant decide if domainNamePrefix or SecretNamePrefix is better
	//if ms.Extra.Ingress.SecretNamePrefix == "" {
	//	utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.secretNamePrefix")
	//	return
	//}

	if ms.Extra.Ingress.Host == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.host")
		return
	}

	namespace := fmt.Sprintf("application-%s", application.ID)
	err = s.simpleRepo.Create(namespace, tenant, application, ingress, ms)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO this could be an event
	// TODO this should be decoupled
	err = s.gitRepo.SaveMicroservice(
		ms.Dolittle.TenantID,
		ms.Dolittle.ApplicationID,
		ms.Environment,
		ms.Dolittle.MicroserviceID,
		ms,
	)

	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, ms)
}
