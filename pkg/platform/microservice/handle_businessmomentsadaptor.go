package microservice

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
)

func (s *handler) handleBusinessMomentsAdaptor(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application, customerTenants []platform.CustomerTenantInfo) {
	// Function assumes access check has taken place
	var ms platform.HttpInputBusinessMomentAdaptorInfo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		utils.RespondWithStatusError(responseWriter, statusErr)
		return
	}

	err := s.businessMomentsAdaptorRepo.Create(msK8sInfo.Namespace, msK8sInfo.Customer, msK8sInfo.Application, customerTenants, ms)
	if statusErr != nil {
		// TODO change
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, statusErr.Error())
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

// TODO notes from talking with Gøran
// acl stuff later look more at CanModifyApplication
func (s *handler) BusinessMomentsAdaptorSave(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)
	customerID := r.Header.Get("Tenant-ID")

	dnsSRV, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	url := fmt.Sprintf("%s/save", strings.TrimSuffix(dnsSRV, "/"))

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":         "TODO businessmomentsadaptor save",
		"namespace":       namespace,
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"customer_id":     customerID,
		"save_url":        url,
	})
}

func (s *handler) BusinessMomentsAdaptorRawData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)
	customerID := r.Header.Get("Tenant-ID")

	dnsSRV, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	url := fmt.Sprintf("%s/rawdata", strings.TrimSuffix(dnsSRV, "/"))

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":         "TODO businessmomentsadaptor get rawdata",
		"namespace":       namespace,
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"customer_id":     customerID,
		"rawDataUrl":      url,
	})
}

func (s *handler) BusinessMomentsAdaptorSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)
	customerID := r.Header.Get("Tenant-ID")

	dnsSRV, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	url := fmt.Sprintf("%s/sync", strings.TrimSuffix(dnsSRV, "/"))

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":         "TODO businessmomentsadaptor sync business moments back to studio",
		"namespace":       namespace,
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"customer_id":     customerID,
		"sync_url":        url,
	})
}
