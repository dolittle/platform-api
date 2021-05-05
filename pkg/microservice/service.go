package microservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

func NewService(k8sClient *kubernetes.Clientset) service {

	return service{
		storage: NewGitStorage(
			"git@github.com:freshteapot/test-deploy-key.git",
			"/tmp/dolittle-k8s",
			"/Users/freshteapot/dolittle/.ssh/test-deploy",
		),
		simpleRepo:      NewSimpleRepo(k8sClient),
		k8sDolittleRepo: NewK8sRepo(k8sClient),
		k8sClient:       k8sClient,
	}
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	var input HttpMicroserviceBase
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

	// TODO Hardcoding to break dev environments
	if input.Environment != "Dev" {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Currently locked down to environment Dev",
		})
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

	switch input.Kind {
	case Simple:
		var ms HttpInputSimpleInfo
		err = json.Unmarshal(b, &ms)
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

		domainPrefix := "freshteapot-taco"
		ingress := k8s.Ingress{
			Host:       fmt.Sprintf("%s.dolittle.cloud", domainPrefix),
			SecretName: fmt.Sprintf("%s-certificate", domainPrefix),
		}

		if tenant.ID != ms.Dolittle.TenantID {
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Currently locked down to tenant 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
			})
			return
		}

		if application.ID != ms.Dolittle.ApplicationID {
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Currently locked down to applicaiton 11b6cf47-5d9f-438f-8116-0d9828654657",
			})
			return
		}

		namespace := fmt.Sprintf("application-%s", application.ID)
		err := s.simpleRepo.Create(namespace, tenant, application, ingress, ms)
		if err != nil {
			// TODO change
			utils.RespondWithJSON(w, http.StatusInternalServerError, err)
			return
		}

		// TODO this could be an event
		// TODO this should be decoupled
		storageBytes, _ := json.Marshal(ms)
		s.storage.Write(ms.Dolittle, storageBytes)
		if err != nil {
			// TODO change
			utils.RespondWithJSON(w, http.StatusInternalServerError, err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, ms)
		return
	case BusinessMomentsAdaptor:
		utils.RespondWithJSON(w, http.StatusOK, "Todo")
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
	microserviceID := vars["microserviceID"]

	data, err := s.storage.Read(HttpInputDolittle{
		TenantID:       tenant.ID,
		ApplicationID:  applicationID,
		MicroserviceID: microserviceID,
	})

	if err != nil {
		// TODO change
		utils.RespondWithJSON(w, http.StatusInternalServerError, err)
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

	data, err := s.storage.GetAll(tenant.ID, applicationID)

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
		"action":          "Remove microservicce",
	})
}
