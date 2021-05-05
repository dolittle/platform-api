package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/thoas/go-funk"
)

func NewService() service {
	return service{
		storage: NewGitStorage(
			"git@github.com:freshteapot/test-deploy-key.git",
			"/tmp/dolittle-k8s",
			"/Users/freshteapot/dolittle/.ssh/test-deploy",
		),
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

	tenantID := "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3"

	storageBytes, err := s.storage.Read(tenantID, applicationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Not able to find application in the storage")
		return
	}

	var application Application
	json.Unmarshal(storageBytes, &application)

	// TODO lookup for environment
	exists := funk.Contains(application.Environments, func(environment HttpInputEnvironment) bool {
		return environment.Name == input.Name || environment.DomainPrefix == input.DomainPrefix
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
