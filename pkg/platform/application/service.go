package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/thoas/go-funk"
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
		k8sDolittleRepo: platform.NewK8sRepo(k8sClient),
		k8sClient:       k8sClient,
	}
}

func (s *service) SaveEnvironment(w http.ResponseWriter, r *http.Request) {
	var input HttpInputEnvironment
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// TODO lookup application
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	fmt.Printf("Lookup %s", applicationID)

	applicationInfo, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Application %s not found", applicationID))
		return
	}

	tenantID := applicationInfo.Tenant.ID

	storageBytes, err := s.storage.Read(tenantID, applicationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Not able to find application in the storage")
		return
	}

	var application Application
	json.Unmarshal(storageBytes, &application)

	// TODO this is not going to work with custom domains.
	// Simple logic to make sure the domainPrefix is not used
	// This is not great and should be linked to actual domains
	exists := funk.Contains(application.Environments, func(environment HttpInputEnvironment) bool {
		found := false
		if environment.Name == input.Name {
			found = true
		}
		if environment.DomainPrefix == input.DomainPrefix {
			found = true
		}
		return found
	})

	if exists {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Environment %s already exists", input.Name))
		return
	}

	application.Environments = append(application.Environments, input)
	storageBytes, _ = json.Marshal(application)
	err = s.storage.Write(tenantID, applicationID, storageBytes)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to storage")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	var input HttpInputApplication
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	_, err = s.storage.Read(input.TenantID, input.ID)
	if err == nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Application already exists")
		return
	}

	// TODO this will overwrite
	application := Application{
		ID:       input.ID,
		Name:     input.Name,
		TenantID: input.TenantID,
	}

	storageBytes, _ := json.Marshal(application)
	err = s.storage.Write(input.TenantID, input.ID, storageBytes)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to storage")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) GetLiveApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantID"]

	// TODO get tenant from syncing the terraform output into the repo (which we might have access to if we use the same repo)
	tenant := k8s.Tenant{
		ID:   tenantID,
		Name: "Customer-Chris",
	}

	if tenant.ID != "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3" {
		utils.RespondWithError(w, http.StatusBadRequest, "Currently locked down to tenant 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3")
		return
	}

	applications, err := s.k8sDolittleRepo.GetApplicationsByTenantID(tenantID)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := HttpResponseApplications{
		ID:           tenantID,
		Name:         tenant.Name,
		Applications: applications,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
