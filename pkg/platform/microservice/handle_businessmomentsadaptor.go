package microservice

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
)

func (s *service) handleBusinessMomentsAdaptor(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place
	var ms platform.HttpInputBusinessMomentAdaptorInfo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		utils.RespondWithStatusError(responseWriter, statusErr)
		return
	}
	ingress := CreateTodoIngress()

	err := s.businessMomentsAdaptorRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms)
	if statusErr != nil {
		// TODO change
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, statusErr.Error())
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

// TODO notes from talking with Gøran
// acl stuff later look more at CanModifyApplication
func (s *service) BusinessMomentsAdaptorSave(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)
	tenantID := r.Header.Get("Tenant-ID")

	dnsSRV, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	url := fmt.Sprintf("%s/save", strings.TrimSuffix(dnsSRV, "/"))

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":        "TODO businessmomentsadaptor save",
		"namespace":      namespace,
		"applicationID":  applicationID,
		"microserviceID": microserviceID,
		"tenantID":       tenantID,
		"saveUrl":        url,
	})
}

func (s *service) BusinessMomentsAdaptorRawData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)
	tenantID := r.Header.Get("Tenant-ID")

	dnsSRV, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	url := fmt.Sprintf("%s/rawdata", strings.TrimSuffix(dnsSRV, "/"))

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":        "TODO businessmomentsadaptor get rawdata",
		"namespace":      namespace,
		"applicationID":  applicationID,
		"microserviceID": microserviceID,
		"tenantID":       tenantID,
		"rawDataUrl":     url,
	})
}

func (s *service) BusinessMomentsAdaptorSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)
	tenantID := r.Header.Get("Tenant-ID")

	dnsSRV, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	url := fmt.Sprintf("%s/sync", strings.TrimSuffix(dnsSRV, "/"))

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":        "TODO businessmomentsadaptor sync business moments back to studio",
		"namespace":      namespace,
		"applicationID":  applicationID,
		"microserviceID": microserviceID,
		"tenantID":       tenantID,
		"syncUrl":        url,
	})
}
