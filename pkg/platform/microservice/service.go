package microservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/thoas/go-funk"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func NewService(gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo, k8sClient *kubernetes.Clientset) service {
	return service{
		gitRepo:                    gitRepo,
		simpleRepo:                 NewSimpleRepo(k8sClient),
		businessMomentsAdaptorRepo: NewBusinessMomentsAdaptorRepo(k8sClient),
		k8sDolittleRepo:            k8sDolittleRepo,
	}
}

// TODO https://dolittle.freshdesk.com/a/tickets/1352 how to add multiple entries to ingress
func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	var input platform.HttpMicroserviceBase
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	userID := r.Header.Get("User-ID")
	applicationID := input.Dolittle.ApplicationID
	environment := input.Environment

	applicationInfo, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Application %s not found", applicationID),
		})
		return
	}

	tenant := k8s.Tenant{
		ID:   applicationInfo.Tenant.ID,
		Name: applicationInfo.Tenant.Name,
	}

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenant.ID, applicationID, userID)
	if !allowed {
		return
	}

	if !s.gitRepo.IsAutomationEnabled(tenant.ID, applicationID, environment) {
		utils.RespondWithError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf(
				"Tenant %s with application %s in environment %s does not allow changes via Studio",
				tenant.ID,
				applicationID,
				environment,
			),
		)
		return
	}

	switch input.Kind {
	case platform.Simple:
		var ms platform.HttpInputSimpleInfo
		err = json.Unmarshal(b, &ms)
		if err != nil {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
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
			utils.RespondWithError(w, http.StatusInternalServerError, "Currently locked down to applicaiton 11b6cf47-5d9f-438f-8116-0d9828654657")
			return
		}

		// TODO I cant decide if domainNamePrefix or SecretNamePrefix is better
		//if ms.Extra.Ingress.SecretNamePrefix == "" {
		//	utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.secretNamePrefix")
		//	return
		//}

		if ms.Extra.Ingress.Host == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.host")
			return
		}

		namespace := fmt.Sprintf("application-%s", application.ID)
		err := s.simpleRepo.Create(namespace, tenant, application, ingress, ms)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// TODO this could be an event
		// TODO this should be decoupled
		storageBytes, _ := json.Marshal(ms)
		err = s.gitRepo.SaveMicroservice(
			ms.Dolittle.TenantID,
			ms.Dolittle.ApplicationID,
			ms.Environment,
			ms.Dolittle.MicroserviceID,
			storageBytes,
		)

		if err != nil {
			// TODO change
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, ms)
		return
	case platform.BusinessMomentsAdaptor:
		s.handleBusinessMomentsAdaptor(w, r, b, applicationInfo)
		return
	default:
		utils.RespondWithError(w, http.StatusBadRequest, "Kind not supported")
	}
}

func (s *service) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	tenantInfo, err := s.gitRepo.GetTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenant := k8s.Tenant{
		ID:   tenantInfo.GUID,
		Name: tenantInfo.Name,
	}

	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]

	data, err := s.gitRepo.GetMicroservice(
		tenant.ID,
		applicationID,
		environment,
		microserviceID,
	)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response interface{}
	json.Unmarshal(data, &response)
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetByApplicationID(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	tenantInfo, err := s.gitRepo.GetTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenant := k8s.Tenant{
		ID:   tenantInfo.GUID,
		Name: tenantInfo.Name,
	}

	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	data, err := s.gitRepo.GetMicroservices(tenant.ID, applicationID)

	if err != nil {
		// TODO change
		utils.RespondWithJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, data)
}

func (s *service) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// I feel we shouldn't need namespace
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	if !s.gitRepo.IsAutomationEnabled(tenantID, applicationID, environment) {
		utils.RespondWithError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf(
				"Tenant %s with application %s in environment %s does not allow changes via Studio",
				tenantID,
				applicationID,
				environment,
			),
		)
		return
	}

	errStr := ""
	statusCode := http.StatusOK

	// Hairy stuff
	msData, err := s.gitRepo.GetMicroservice(tenantID, applicationID, environment, microserviceID)
	if err == nil {
		// LOG THIS
		var whatKind platform.HttpInputMicroserviceKind
		err = json.Unmarshal(msData, &whatKind)
		if err == nil {
			switch whatKind.Kind {
			case platform.Simple:
				err = s.simpleRepo.Delete(namespace, microserviceID)
			case platform.BusinessMomentsAdaptor:
				err = s.businessMomentsAdaptorRepo.Delete(namespace, microserviceID)
			}

			if err != nil {
				statusCode = http.StatusUnprocessableEntity
				errStr = err.Error()
			}
		}
	}

	// TODO need to delete from gitrepo
	err = s.gitRepo.DeleteMicroservice(tenantID, applicationID, environment, microserviceID)
	if err != nil {
		statusCode = http.StatusUnprocessableEntity
		errStr = err.Error()
	}

	utils.RespondWithJSON(w, statusCode, map[string]string{
		"namespace":       namespace,
		"error":           errStr,
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"action":          "Remove microservice",
	})
}

func (s *service) GetLiveByApplicationID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	application, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	microservices, err := s.k8sDolittleRepo.GetMicroservices(applicationID)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := platform.HttpResponseMicroservices{
		Application: platform.ShortInfo{
			Name: application.Name,
			ID:   application.ID,
		},
		Microservices: microservices,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetPodStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := strings.ToLower(vars["environment"])

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	status, err := s.k8sDolittleRepo.GetPodStatus(applicationID, microserviceID, environment)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//response := HttpResponsePodStatus{
	//	Application: platform.ShortInfo{
	//		Name: application.Name,
	//		ID:   application.ID,
	//	},
	//	Microservice: platform.ShortInfoWithEnvironment{
	//		Name: application.Name,
	//		ID:   application.ID,
	//	},
	//}

	utils.RespondWithJSON(w, http.StatusOK, status)
}

func (s *service) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	podName := vars["podName"]
	containerName := r.FormValue("containerName")
	// TODO how to ignore?
	if containerName == "" {
		containerName = "head"
	}

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	logData, err := s.k8sDolittleRepo.GetLogs(applicationID, containerName, podName)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, platform.HttpResponsePodLog{
		ApplicationID:  applicationID,
		MicroserviceID: "TODO",
		PodName:        podName,
		Logs:           logData,
	})
}

func (s *service) GetConfigMap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	configMapName := vars["configMapName"]

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/json"
	}

	filterContentType := funk.ContainsString([]string{
		"application/json",
		"application/yaml",
	}, contentType)

	if !filterContentType {
		utils.RespondWithError(w, http.StatusBadRequest, "Content-Type header not supported")
		return
	}

	// Hmm this will let them see things they are not allowed to see.
	// But it wont let them update it
	// TODO when / if we allow update, we will need more protection
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	configMap, err := s.k8sDolittleRepo.GetConfigMap(applicationID, configMapName)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Config map %s not found in application %s", configMapName, applicationID))
		return
	}

	// Little hack to work around https://github.com/kubernetes/client-go/issues/861
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"

	output, _ := yaml.Marshal(configMap)

	switch contentType {
	case "application/json":
		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"json": configMap,
			"yaml": string(output),
		})
	case "application/yaml":
		utils.RespondWithYAML(w, http.StatusOK, output)
	}

}

func (s *service) CanI(w http.ResponseWriter, r *http.Request) {
	type httpInput struct {
		UserID        string `json:"user_id"`
		TenantID      string `json:"tenant_id"`
		ApplicationID string `json:"application_id"`
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var input httpInput
	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	allowed, err := s.k8sDolittleRepo.CanModifyApplication(input.TenantID, input.ApplicationID, input.UserID)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allowedStr := "allowed"
	if !allowed {
		allowedStr = "denied"
	}
	utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{
		"allowed": allowedStr,
	})
}
