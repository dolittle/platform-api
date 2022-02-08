package microservice

import (
	_ "embed"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/thoas/go-funk"
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

	// TODO why is this broken?
	pathExists := funk.Contains(applicationInfo.Ingresses, func(info platform.Ingress) bool {
		return info.Path == ms.Extra.Ingress.Path
	})

	if pathExists {
		utils.RespondWithError(w, http.StatusBadRequest, "ms.Extra.Ingress.Path The path is already in use")
		return
	}

	err := s.simpleRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, customerTenants, ms)
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
