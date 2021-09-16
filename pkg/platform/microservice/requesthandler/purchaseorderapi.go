package requesthandler

import (
	_ "embed"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
)

type purchaseOrderApiHandler struct {
	parser  Parser
	repo    purchaseorderapi.Repo
	gitRepo storage.Repo
}

func NewPurchaseOrderApiHandler(parser Parser, repo purchaseorderapi.Repo, gitRepo storage.Repo) Handler {
	return &purchaseOrderApiHandler{parser, repo, gitRepo}
}

func (s *purchaseOrderApiHandler) Create(request *http.Request, inputBytes []byte, applicationInfo platform.Application) (platform.Microservice, *Error) {
	// Function assumes access check has taken place
	var ms platform.HttpInputPurchaseOrderInfo
	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		return nil, parserError
	}

	tenant, err := getFirstTenant(s.gitRepo, applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		return nil, NewInternalError(err)
	}

	err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		return nil, NewInternalError(err)
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
		return nil, NewInternalError(err)
	}

	return ms, nil
}

func (s *purchaseOrderApiHandler) Delete(namespace string, microserviceID string) *Error {
	if err := s.repo.Delete(namespace, microserviceID); err != nil {
		return NewInternalError(err)
	}
	return nil
}
