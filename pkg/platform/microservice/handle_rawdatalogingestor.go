package microservice

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/thoas/go-funk"
)

func (s *handler) handleRawDataLogIngestor(
	responseWriter http.ResponseWriter,
	r *http.Request,
	inputBytes []byte,
	applicationInfo platform.Application,
	customerTenants []platform.CustomerTenantInfo,
) {
	// Function assumes access check has taken place
	var ms platform.HttpInputRawDataLogIngestorInfo

	// TODO: Fix this when we get to work on rawDataLogIngestorRepo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		utils.RespondWithStatusError(responseWriter, statusErr)
		return
	}
	// TODO changing writeTo will break this.
	// TODO does this exist?
	if ms.Extra.WriteTo == "" {
		ms.Extra.WriteTo = "nats"
	}

	writeToCheck := funk.Contains([]string{
		"stdout",
		"nats",
	}, func(filter string) bool {
		return strings.HasSuffix(ms.Extra.WriteTo, filter)
	})

	if !writeToCheck {
		utils.RespondWithError(responseWriter, http.StatusForbidden, "writeTo is not valid, leave empty or set to stdout")
		return
	}

	exists, _, err := s.rawDataLogIngestorRepo.Exists(msK8sInfo.Namespace, ms.Environment)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		// Create in Kubernetes
		err = s.rawDataLogIngestorRepo.Create(msK8sInfo.Namespace, msK8sInfo.Customer, msK8sInfo.Application, customerTenants, ms) //TODO:
	} else {
		err = s.rawDataLogIngestorRepo.Update(msK8sInfo.Namespace, msK8sInfo.Customer, msK8sInfo.Application, ms) //TODO:
	}

	if err != nil {
		// TODO change
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
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
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(responseWriter, http.StatusOK, ms)
}
