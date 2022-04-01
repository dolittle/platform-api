package microservice

import (
	_ "embed"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"k8s.io/apimachinery/pkg/util/validation"
)

func (s *service) handleSimpleMicroservice(
	w http.ResponseWriter,
	r *http.Request,
	inputBytes []byte,
	applicationInfo platform.Application,
	customerTenants []platform.CustomerTenantInfo,
) {
	// Function assumes access check has taken place

	var ms platform.HttpInputSimpleInfo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		utils.RespondWithStatusError(w, statusErr)
		return
	}

	if ms.Extra.Ispublic {
		if CheckIfIngressPathInUseInEnvironment(applicationInfo.Ingresses, ms.Environment, ms.Extra.Ingress.Path) {
			utils.RespondWithError(w, http.StatusBadRequest, "ms.Extra.Ingress.Path The path is already in use")
			return
		}
	}

	// If 0, let it default to port 80
	if ms.Extra.HeadPort == 0 {
		ms.Extra.HeadPort = 80
	}

	if validation.IsValidPortNum(int(ms.Extra.HeadPort)) != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ms.Extra.HeadPort not a valid port number")
		return
	}

	err := s.simpleRepo.Create(msK8sInfo.Namespace, msK8sInfo.Customer, msK8sInfo.Application, customerTenants, ms)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO this could be an event
	// TODO this should be decoupled
	err = s.gitRepo.SaveMicroservice(
		ms.Dolittle.CustomerID,
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
