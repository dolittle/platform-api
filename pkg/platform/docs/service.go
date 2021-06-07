package docs

import (
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

func NewService(gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo) service {
	return service{
		gitRepo:         gitRepo,
		k8sDolittleRepo: k8sDolittleRepo,
	}
}

func (s *service) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	//
	//// TODO get tenant from syncing the terraform output into the repo (which we might have access to if we use the same repo)
	terraformCustomer, err := s.gitRepo.GetTenant(tenantID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	terraformCustomer, err := s.gitRepo.GetTenant(tenantID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//microservices, err := s.gitRepo.GetMicroservices(tenantID, applicationID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"customer":    terraformCustomer,
		"application": terraformApplication,
	})
}
