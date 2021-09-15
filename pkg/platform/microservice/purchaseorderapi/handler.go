package purchaseorderapi

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/requests"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

type RequestHandler struct {
	parser  parser.Parser
	repo    Repo
	gitRepo storage.Repo
}

func NewRequestHandler(parser parser.Parser, repo Repo, gitRepo storage.Repo) requests.RequestHandler {
	return &RequestHandler{parser, repo, gitRepo}
}

func (s *RequestHandler) Create(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) error {
	// Function assumes access check has taken place
	var ms platform.HttpInputPurchaseOrderInfo
	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		utils.RespondWithStatusError(responseWriter, parserError)
		return parserError
	}
	tenant, err := s.getConfiguredTenant(applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}

	exists, err := s.repo.Exists(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}
	if exists {
		utils.RespondWithError(responseWriter, http.StatusConflict, fmt.Sprintf("A Purchase Order API Microservice with ID %s already exists in %s enironment in application %s under customer %s", ms.Dolittle.MicroserviceID, ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
		return nil
	}

	err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}
	err = s.gitRepo.SaveMicroservice(
		ms.Dolittle.TenantID,
		ms.Dolittle.ApplicationID,
		ms.Environment,
		ms.Dolittle.MicroserviceID,
		ms,
	)

	if err != nil {
		// TODO change
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}

	utils.RespondWithJSON(responseWriter, http.StatusOK, ms)
	return nil
}

func (s *RequestHandler) Delete(namespace string, microserviceID string) error {
	return s.repo.Delete(namespace, microserviceID)
}

func (s *RequestHandler) getConfiguredTenant(customerID, appplicationID, environment string) (platform.TenantId, error) {
	application, err := s.gitRepo.GetApplication(customerID, appplicationID)
	if err != nil {
		return "", err
	}
	return application.GetTenantForEnvironment(environment)
}
