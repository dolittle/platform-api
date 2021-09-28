package purchaseorderapi

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/requests"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/google/uuid"
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

// Create creates a new PurchaseOrderAPI microservice and creates a RawDataLog microservice too if it didn't already exist
func (s *RequestHandler) Create(responseWriter http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) error {
	// Function assumes access check has taken place

	logger := s.logContext.WithFields(logrus.Fields{
		"handler": "PurchaseOrderAPI",
		"method":  "Create",
	})

	var ms platform.HttpInputPurchaseOrderInfo
	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		logger.WithError(parserError).Error("Failed to parse input")
		utils.RespondWithStatusError(responseWriter, parserError)
		return parserError
	}

	logger = logger.WithFields(logrus.Fields{
		"tenantID":      applicationInfo.Tenant.ID,
		"applicationID": applicationInfo.ID,
		"environment":   ms.Environment,
	})
	logger.Debug("Starting to create a PurchaseOrderAPI microservice")

	tenant, err := s.getConfiguredTenant(applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to get configured tenant")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}

	microservices, err := s.gitRepo.GetMicroservices(msK8sInfo.Tenant.ID, msK8sInfo.Application.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to get microservices from GitRepo")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}
	for _, microservice := range microservices {
		if microservice.Kind == platform.MicroserviceKindPurchaseOrderAPI && strings.EqualFold(microservice.Environment, ms.Environment) {
			logger.Warn("A Purchase Order API Microservice already exists in GitRepo")
			utils.RespondWithError(responseWriter, http.StatusConflict, fmt.Sprintf("A Purchase Order API Microservice already exists in %s environment in application %s under customer %s", ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
			return nil
		}
	}

	exists, err := s.purchaseOrderAPIExists(responseWriter, msK8sInfo, ms, tenant, logger)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = s.ensureRawDataLogExists(responseWriter, msK8sInfo, ms, logger)
	if err != nil {
		return err
	}
	// TODO: Since we only do ingress validation on the creation of RawDataLog, this means that if it exists - we don't do any validation.

	err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		logger.WithError(err).Error("Failed to create Purchase Order API")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return nil
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
		logger.WithError(err).Error("Failed to save Purchase Order API in GitRepo")
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

func (s *RequestHandler) purchaseOrderAPIExists(responseWriter http.ResponseWriter, msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, tenant platform.TenantId, logger *logrus.Entry) (exists bool, err error) {
	exists, err = s.repo.EnvironmentHasPurchaseOrderAPI(msK8sInfo.Namespace, ms)
	if err != nil {
		logger.WithError(err).Error("Failed to check if environment has Purchase Order API with K8sRepo")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}
	if exists {
		logger.Warn("A Purchase Order API Microservice already exists in K8sRepo")
		utils.RespondWithError(responseWriter, http.StatusConflict, fmt.Sprintf("A Purchase Order API Microservice already exists in %s environment in application %s under customer %s", ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
		return
	}

	exists, err = s.repo.Exists(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		logger.WithError(err).Error("Failed to check if Purchase Order API exists with K8sRepo")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}
	if exists {
		logger.WithField("microserviceID", ms.Dolittle.MicroserviceID).Warn("A Purchase Order API Microservice with the same name already exists in K8sRepo")
		utils.RespondWithError(responseWriter, http.StatusConflict, fmt.Sprintf("A Purchase Order API Microservice with ID %s already exists in %s environment in application %s under customer %s", ms.Dolittle.MicroserviceID, ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
		return
	}
	return
}

func (s *RequestHandler) ensureRawDataLogExists(responseWriter http.ResponseWriter, msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, logger *logrus.Entry) error {
	rawDataLogExists, err := s.rawdatalogRepo.Exists(msK8sInfo.Namespace, ms.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to check if Raw Data Log exists")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}
	if !rawDataLogExists {
		logger.Debug("Raw Data Log does not exist, creating a new one")
		err = s.createRawDataLog(responseWriter, msK8sInfo, ms, logger)
		if err != nil {
			return err
		}
	} else {
		err = s.rawdatalogRepo.Update(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, s.extractRawDataLogInfo(ms))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RequestHandler) createRawDataLog(responseWriter http.ResponseWriter, msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, logger *logrus.Entry) error {
	input := s.extractRawDataLogInfo(ms)
	err := s.rawdatalogRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, input)
	if err != nil {
		logger.WithError(err).Error("Failed to create Raw Data Log")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}
	err = s.gitRepo.SaveMicroservice(
		input.Dolittle.TenantID,
		input.Dolittle.ApplicationID,
		input.Environment,
		input.Dolittle.MicroserviceID,
		input)
	if err != nil {
		logger.WithError(err).Error("Failed to save Raw Data Log in GitRepo")
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}
	return nil
}

func (s *RequestHandler) extractRawDataLogInfo(ms platform.HttpInputPurchaseOrderInfo) platform.HttpInputRawDataLogIngestorInfo {
	return platform.HttpInputRawDataLogIngestorInfo{
		MicroserviceBase: platform.MicroserviceBase{
			Name:        ms.Extra.RawDataLogName,
			Kind:        platform.MicroserviceKindRawDataLogIngestor,
			Environment: ms.Environment,
			Dolittle: platform.HttpInputDolittle{
				ApplicationID:  ms.Dolittle.ApplicationID,
				TenantID:       ms.Dolittle.TenantID,
				MicroserviceID: uuid.New().String(),
			},
		},
		Extra: platform.HttpInputRawDataLogIngestorExtra{
			// TODO these images won't evolve automatically
			Headimage:    "dolittle/platform-api:latest",
			Runtimeimage: "dolittle/runtime:6.1.0",
			Ingress:      ms.Extra.Ingress,
			WriteTo:      "nats",
			Webhooks:     ms.Extra.Webhooks,
		},
	}
}
