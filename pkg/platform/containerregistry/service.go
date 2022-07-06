package containerregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	applicationK8s "github.com/dolittle/platform-api/pkg/platform/application/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"

	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type HTTPResponseImages struct {
	Url    string   `json:"url"`
	Images []string `json:"images"`
}

type HTTPResponseTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type ImageTag struct {
	Name         string    `json:"name"`
	LastModified time.Time `json:"lastModified"`
}

type HTTPResponseTags2 struct {
	Name string     `json:"name"`
	Tags []ImageTag `json:"tags"`
}

// TODO maybe tags returns array of objects
type ContainerRegistryCredentials struct {
	URL      string
	Username string
	Password string
}

type ContainerRegistryRepo interface {
	GetImages(credentials ContainerRegistryCredentials) ([]string, error)
	GetTags(credentials ContainerRegistryCredentials, image string) ([]string, error)
	GetTags2(credentials ContainerRegistryCredentials, image string) ([]ImageTag, error)
}

type service struct {
	gitRepo         storage.Repo
	repo            ContainerRegistryRepo
	k8sDolittleRepo platformK8s.K8sPlatformRepo
	logContext      logrus.FieldLogger
}

func NewService(
	gitRepo storage.Repo,
	repo ContainerRegistryRepo,
	k8sDolittleRepo platformK8s.K8sPlatformRepo,
	logContext logrus.FieldLogger,
) service {
	return service{
		gitRepo:         gitRepo,
		repo:            repo,
		k8sDolittleRepo: k8sDolittleRepo,
		logContext:      logContext,
	}
}

func (s *service) getContainerRegistryCredentialsFromKubernetes(logContext logrus.FieldLogger, applicationID string, containerRegistryName string) (ContainerRegistryCredentials, error) {
	secretName := "acr"
	secret, err := s.k8sDolittleRepo.GetSecret(logContext, applicationID, secretName)
	if err != nil {
		logContext.WithField("error", err).Error("failed to talk to kubernetes")
		return ContainerRegistryCredentials{}, err
	}

	data := secret.Data[".dockerconfigjson"]

	var config applicationK8s.DockerConfigJSON
	_ = json.Unmarshal(data, &config)

	containerRegistryKey := fmt.Sprintf("%s.azurecr.io", containerRegistryName)

	containerRegistryCredentials, ok := config.Auths[containerRegistryKey]

	if !ok {
		logContext.WithField("error", err).Error("acr is missing")
		return ContainerRegistryCredentials{}, err
	}

	credentials := ContainerRegistryCredentials{
		URL:      fmt.Sprintf("https://%s", containerRegistryKey),
		Username: containerRegistryCredentials.Username,
		Password: containerRegistryCredentials.Password,
	}

	return credentials, nil
}

func (s *service) GetImages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	customer, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "GetImages",
		"customerID":    customerID,
		"applicationID": applicationID,
		"userID":        userID,
	})

	credentials, err := s.getContainerRegistryCredentialsFromKubernetes(logContext, applicationID, customer.ContainerRegistryName)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	images, err := s.repo.GetImages(credentials)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get images")
		return
	}

	response := HTTPResponseImages{
		Url:    fmt.Sprintf("%s.azurecr.io", customer.ContainerRegistryName),
		Images: images,
	}

	utils.RespondWithJSON(
		w,
		http.StatusOK,
		response,
	)
}

func (s *service) GetTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	imageName := vars["imageName"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	customer, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "GetTags",
		"customerID":    customerID,
		"applicationID": applicationID,
		"userID":        userID,
	})

	credentials, err := s.getContainerRegistryCredentialsFromKubernetes(logContext, applicationID, customer.ContainerRegistryName)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	tags, err := s.repo.GetTags(credentials, imageName)
	if err != nil {
		if err == ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, "Tag was not found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get tags")
		return
	}

	response := HTTPResponseTags{
		Name: imageName,
		Tags: tags,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetTags2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	imageName := vars["imageName"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	customer, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "GetTags",
		"customerID":    customerID,
		"applicationID": applicationID,
		"userID":        userID,
	})

	credentials, err := s.getContainerRegistryCredentialsFromKubernetes(logContext, applicationID, customer.ContainerRegistryName)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	tags, err := s.repo.GetTags2(credentials, imageName)
	if err != nil {
		if err == ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, "Tag was not found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get tags")
		return
	}

	response := HTTPResponseTags2{
		Name: imageName,
		Tags: tags,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
