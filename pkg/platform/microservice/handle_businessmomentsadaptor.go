package microservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
)

func (s *service) handleBusinessMomentsAdaptor(w http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place
	var ms platform.HttpInputBusinessMomentAdaptorInfo
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Application id in uri does not match application id in body")
		return
	}

	namespace := fmt.Sprintf("application-%s", application.ID)
	err = s.businessMomentsAdaptorRepo.Create(namespace, tenant, application, ingress, ms)
	if err != nil {
		// TODO change
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
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, ms)
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
