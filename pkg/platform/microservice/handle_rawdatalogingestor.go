package microservice

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/thoas/go-funk"
)

func (s *service) handleRawDataLogIngestor(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place
	var ms platform.HttpInputRawDataLogIngestorInfo

	// TODO: Fix this when we get to work on rawDataLogIngestorRepo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		utils.RespondWithStatusError(responseWriter, statusErr)
		return
	}
	storedIngress, err := s.getStoredIngress(applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}
	ingress := k8s.Ingress{
		Host:       storedIngress.Host,
		SecretName: storedIngress.SecretName,
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

	// TODO lookup to see if it exists?
	exists := false
	//exists := true s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//exists, err := s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//if err != nil {
	//	// TODO change
	//	utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	//	return
	//}
	if !exists {
		// Create in Kubernetes
		err = s.rawDataLogIngestorRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms) //TODO:
	} else {
		err = s.rawDataLogIngestorRepo.Update(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms) //TODO:
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

func (s *service) getStoredIngress(customerID, applicationID, environment string) (platform.EnvironmentIngress, error) {
	storedIngress := platform.EnvironmentIngress{}
	application, err := s.gitRepo.GetApplication(customerID, applicationID)
	if err != nil {
		// TODO change
		return storedIngress, err
	}
	tenant, err := application.GetTenantForEnvironment(environment)
	if err != nil {
		// TODO change
		return storedIngress, err
	}
	storedIngress, ok := application.Environments[funk.IndexOf(application.Environments, func(e platform.HttpInputEnvironment) bool {
		return e.Name == environment
	})].Ingresses[tenant]
	if !ok {
		return storedIngress, fmt.Errorf("Failed to get stored ingress for tenant %s in environment %s", string(tenant), environment)
	}
	return storedIngress, nil
}
