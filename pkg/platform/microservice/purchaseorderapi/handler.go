package purchaseorderapi

import (
	_ "embed"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/requests"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
)

type RequestHandler struct {
	parser         parser.Parser
	repo           Repo
	gitRepo        storage.Repo
	rawdatalogRepo rawdatalog.RawDataLogIngestorRepo
	logContext     logrus.FieldLogger
}

func NewRequestHandler(parser parser.Parser, repo Repo, gitRepo storage.Repo, rawDataLogIngestorRepo rawdatalog.RawDataLogIngestorRepo, logContext logrus.FieldLogger) requests.RequestHandler {
	return &RequestHandler{parser, repo, gitRepo, rawDataLogIngestorRepo, logContext}
}

func (s *RequestHandler) Create(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) error {
	// Function assumes access check has taken place
	var ms platform.HttpInputPurchaseOrderInfo
	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		utils.RespondWithStatusError(responseWriter, parserError)
		return parserError
	}

	exists, err := s.rawdatalogRepo.Exists(msK8sInfo.Namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return nil
	}

	if !exists {
		// @joel create it!!!
		err := s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ms)
		if err != nil {
			utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
			return nil
		}
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
