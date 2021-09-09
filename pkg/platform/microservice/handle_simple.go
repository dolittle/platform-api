package microservice

import (
	_ "embed"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

func (s *service) handleSimpleMicroservice(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place

	var ms platform.HttpInputSimpleInfo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		utils.RespondWithStatusError(responseWriter, statusErr)
		return
	}

	ingress := CreateIngress()

	// TODO I cant decide if domainNamePrefix or SecretNamePrefix is better
	//if ms.Extra.Ingress.SecretNamePrefix == "" {
	//	utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.secretNamePrefix")
	//	return
	//}

	if ms.Extra.Ingress.Host == "" {
		utils.RespondWithError(responseWriter, http.StatusBadRequest, "Missing extra.ingress.host")
		return
	}

	err := s.simpleRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
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
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(responseWriter, http.StatusOK, ms)
}
