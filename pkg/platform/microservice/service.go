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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
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

	tenantID := r.Header.Get("Tenant-ID")
	userID := r.Header.Get("User-ID")
	if tenantID == "" || userID == "" {
		// If the middleware is enabled this shouldn't happen
		utils.RespondWithError(w, http.StatusForbidden, "Tenant-ID and User-ID is missing from the headers")
		return
	}

	// TODO remove when happy with things
	if tenantID != "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3" || input.Environment != "Dev" {
		utils.RespondWithError(w, http.StatusBadRequest, "Currently locked down to tenant 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3 and environment Dev")
		return
	}

	applicationInfo, err := s.k8sDolittleRepo.GetApplication(input.Dolittle.ApplicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Application %s not found", input.Dolittle.ApplicationID),
		})
		return
	}

	allowed, err := s.k8sDolittleRepo.CanModifyApplication(tenantID, applicationInfo.ID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
		return
	}

	tenant := k8s.Tenant{
		ID:   applicationInfo.Tenant.ID,
		Name: applicationInfo.Tenant.Name,
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

		if tenant.ID != ms.Dolittle.TenantID {
			utils.RespondWithError(w, http.StatusBadRequest, "tenant id in the system doe not match the one in the input")
			return
		}

		if application.ID != ms.Dolittle.ApplicationID {
			utils.RespondWithError(w, http.StatusInternalServerError, "Currently locked down to applicaiton 11b6cf47-5d9f-438f-8116-0d9828654657")
			return
		}

		if ms.Extra.Ingress.SecretNamePrefix == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.secretNamePrefix")
			return
		}

		if ms.Extra.Ingress.Host == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.host")
			return
		}

		ingress := k8s.Ingress{
			Host:       ms.Extra.Ingress.Host,
			SecretName: fmt.Sprintf("%s-certificate", ms.Extra.Ingress.SecretNamePrefix),
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
	vars := mux.Vars(r)

	tenant := k8s.Tenant{
		ID:   "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
		Name: "Customer-Chris",
	}

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
	vars := mux.Vars(r)

	applicationID := vars["applicationID"]

	tenant := k8s.Tenant{
		ID:   "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
		Name: "Customer-Chris",
	}

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
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)

	err := s.simpleRepo.Delete(namespace, microserviceID)
	fmt.Println("err", err)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"namespace":       namespace,
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"action":          "Remove microservice",
	})
}

func (s *service) GetLiveByApplicationID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
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

	status, err := s.k8sDolittleRepo.GetPodStatus(applicationID, microserviceID, environment)
	if err != nil {
		// TODO change
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
