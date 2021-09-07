package microservice

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/thoas/go-funk"
)

func (s *service) handleRawDataLogIngestor(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place
	var ms platform.HttpInputRawDataLogIngestorInfo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		utils.RespondWithStatusError(responseWriter, statusErr)
		return
	}
	ingress := createIngress()

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

	// TODO lookup to see if it exists?
	exists := false
	//exists := true s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//exists, err := s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//if err != nil {
	//	// TODO change
	//	utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	//	return
	//}
	var err error
	if !exists {
		// Create in Kubernetes
		err = s.rawDataLogIngestorRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms)
	} else {
		err = s.rawDataLogIngestorRepo.Update(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms)
	}

	if err != nil {
		// TODO change
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
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(responseWriter, http.StatusOK, ms)
}
