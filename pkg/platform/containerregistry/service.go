package containerregistry

import (
	"fmt"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type HTTPResponseImages struct {
	Images []string `json:"images"`
}

type HTTPResponseTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
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
}

type service struct {
	gitRepo    storage.Repo
	repo       ContainerRegistryRepo
	logContext logrus.FieldLogger
}

func NewService(
	gitRepo storage.Repo,
	repo ContainerRegistryRepo,
	logContext logrus.FieldLogger,
) service {
	return service{
		gitRepo:    gitRepo,
		repo:       repo,
		logContext: logContext,
	}
}

func (s *service) GetImages(w http.ResponseWriter, r *http.Request) {

	customerID := r.Header.Get("Tenant-ID")
	// TODO might need to confirm access

	customer, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	credentials := ContainerRegistryCredentials{
		URL:      fmt.Sprintf("%s.azurecr.io", customer.ContainerRegistryName),
		Username: customer.ContainerRegistryUsername,
		Password: customer.ContainerRegistryPassword,
	}

	images, err := s.repo.GetImages(credentials)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get images")
		return
	}

	response := HTTPResponseImages{
		Images: images,
	}

	utils.RespondWithJSON(
		w,
		http.StatusOK,
		response,
	)
}

func (s *service) GetTags(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")
	// TODO might need to confirm access

	customer, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	credentials := ContainerRegistryCredentials{
		URL:      fmt.Sprintf("%s.azurecr.io", customer.ContainerRegistryName),
		Username: customer.ContainerRegistryUsername,
		Password: customer.ContainerRegistryPassword,
	}

	vars := mux.Vars(r)
	imageName := vars["imageName"]

	tags, err := s.repo.GetTags(credentials, imageName)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get tags")
		return
	}

	response := HTTPResponseTags{
		Name: imageName,
		Tags: tags,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
