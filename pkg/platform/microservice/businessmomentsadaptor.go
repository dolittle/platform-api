package microservice

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
)

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
