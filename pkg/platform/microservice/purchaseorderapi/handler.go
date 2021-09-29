package purchaseorderapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Error struct {
	StatusCode int
	Err        error
}

func newInternalError(err error) *Error {
	return &Error{http.StatusInternalServerError, err}
}
func newBadRequest(err error) *Error {
	return &Error{http.StatusBadRequest, err}
}
func newForbidden(err error) *Error {
	return &Error{http.StatusForbidden, err}
}
func newConflict(err error) *Error {
	return &Error{http.StatusConflict, err}
}

func (e *Error) Error() string {
	return fmt.Sprintf("status %d: Err %v", e.StatusCode, e.Err)
}

type Handler struct {
	parser         parser.Parser
	repo           Repo
	gitRepo        storage.Repo
	rawdatalogRepo rawdatalog.RawDataLogIngestorRepo
	logContext     logrus.FieldLogger
}

func NewHandler(parser parser.Parser, repo Repo, gitRepo storage.Repo, rawDataLogIngestorRepo rawdatalog.RawDataLogIngestorRepo, logContext logrus.FieldLogger) *Handler {
	return &Handler{parser, repo, gitRepo, rawDataLogIngestorRepo, logContext}
}

// Create creates a new PurchaseOrderAPI microservice and creates a RawDataLog microservice too if it didn't already exist
func (s *Handler) Create(inputBytes []byte, applicationInfo platform.Application) (platform.HttpInputPurchaseOrderInfo, *Error) {
	// Function assumes access check has taken place
	var ms platform.HttpInputPurchaseOrderInfo
	logger := s.logContext.WithFields(logrus.Fields{
		"handler": "PurchaseOrderAPI",
		"method":  "Create",
	})

	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		logger.WithError(parserError).Error("Failed to parse input")
		return ms, newBadRequest(fmt.Errorf("failed to parse input: %w", parserError))
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
		return ms, newInternalError(fmt.Errorf("failed to get configured tenant: %w", err))
	}

	if statusErr := s.ensurePurchaseOrderAPIDoesNotExist(msK8sInfo, ms, tenant, logger); statusErr != nil {
		return ms, statusErr
	}

	if statusErr := s.ensureRawDataLogExists(msK8sInfo, ms, logger); statusErr != nil {
		return ms, statusErr
	}

	return ms, s.createPurchaseOrderAPI(msK8sInfo, ms, tenant, logger)
}

// Update updates an existing PurchaseOrderAPI microservice and creates a RawDataLog microservice too if it didn't already exist
func (s *Handler) UpdateWebhooks(inputBytes []byte, applicationInfo platform.Application) (platform.HttpInputPurchaseOrderInfo, *Error) {
	// Function assumes access check has taken place
	var ms platform.HttpInputPurchaseOrderInfo
	logger := s.logContext.WithFields(logrus.Fields{
		"handler": "PurchaseOrderAPI",
		"method":  "UpdateWebhooks",
	})

	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		logger.WithError(parserError).Error("Failed to parse input")
		return ms, newBadRequest(fmt.Errorf("failed to parse input: %w", parserError))
	}

	logger = logger.WithFields(logrus.Fields{
		"tenantID":      applicationInfo.Tenant.ID,
		"applicationID": applicationInfo.ID,
		"environment":   ms.Environment,
	})
	logger.Debug("Starting to update PurchaseOrderAPI microservice")

	tenant, err := s.getConfiguredTenant(applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to get configured tenant")
		return ms, newInternalError(fmt.Errorf("failed to get configured tenant: %w", err))
	}
	if statusErr := s.ensurePurchaseOrderAPIExists(msK8sInfo, ms, tenant, logger); statusErr != nil {
		return ms, statusErr
	}

	if statusErr := s.ensureRawDataLogExists(msK8sInfo, ms, logger); statusErr != nil {
		return ms, statusErr
	}
	return ms, s.updatePurchaseOrderAPIWebhooks(msK8sInfo, ms.Extra.Webhooks, ms.Environment, ms.Dolittle.MicroserviceID, logger)
}

func (s *Handler) Delete(namespace, microserviceID string) error {
	if err := s.repo.Delete(namespace, microserviceID); err != nil {
		return fmt.Errorf("failed to delete Purchase Order API: %w", err)
	}
	return nil
}

func (s *Handler) getConfiguredTenant(customerID, appplicationID, environment string) (platform.TenantId, error) {
	application, err := s.gitRepo.GetApplication(customerID, appplicationID)
	if err != nil {
		return "", err
	}
	return application.GetTenantForEnvironment(environment)
}

func (s *Handler) ensurePurchaseOrderAPIDoesNotExist(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, tenant platform.TenantId, logger *logrus.Entry) *Error {
	microservices, err := s.gitRepo.GetMicroservices(msK8sInfo.Tenant.ID, msK8sInfo.Application.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to get microservices from GitRepo")
		return newInternalError(fmt.Errorf("failed to get microservices from GitRepo: %w", err))
	}
	for _, microservice := range microservices {
		if microservice.Kind == platform.MicroserviceKindPurchaseOrderAPI && strings.EqualFold(microservice.Environment, ms.Environment) {
			logger.Warn("A Purchase Order API Microservice already exists in GitRepo")
			return newConflict(fmt.Errorf("A Purchase Order API Microservice already exists in %s environment in application %s under customer %s", ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
		}
	}

	exists, err := s.repo.EnvironmentHasPurchaseOrderAPI(msK8sInfo.Namespace, ms)
	if err != nil {
		logger.WithError(err).Error("Failed to check if environment has Purchase Order API with K8sRepo")
		return newInternalError(fmt.Errorf("failed to check if environment has Purchase Order API with K8sRepo: %w", err))
	}
	if exists {
		logger.Warn("A Purchase Order API Microservice already exists in K8sRepo")
		return newConflict(fmt.Errorf("a Purchase Order API Microservice already exists in %s environment in application %s under customer %s", ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
	}

	exists, err = s.repo.Exists(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		logger.WithError(err).Error("Failed to check if Purchase Order API exists with K8sRepo")
		return newInternalError(fmt.Errorf("failed to check if Purchase Order API exists with K8sRepo: %w", err))
	}
	if exists {
		logger.WithField("microserviceID", ms.Dolittle.MicroserviceID).Warn("A Purchase Order API Microservice with the same name already exists in K8sRepo")
		return newConflict(fmt.Errorf("a Purchase Order API Microservice with ID %s already exists in %s environment in application %s under customer %s", ms.Dolittle.MicroserviceID, ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
	}
	return nil
}

func (s *Handler) createPurchaseOrderAPI(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, tenant platform.TenantId, logger *logrus.Entry) *Error {
	if err := s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms); err != nil {
		logger.WithError(err).Error("Failed to create Purchase Order API")
		return newInternalError(fmt.Errorf("failed to create Purchase Order API: %w", err))
	}

	if err := s.gitRepo.SaveMicroservice(ms.Dolittle.TenantID, ms.Dolittle.ApplicationID, ms.Environment, ms.Dolittle.MicroserviceID, ms); err != nil {
		// TODO change
		logger.WithError(err).Error("Failed to save Purchase Order API in GitRepo")
		return newInternalError(fmt.Errorf("failed to save Purchase Order API in GitRepo"))
	}
	return nil
}

func (s *Handler) ensureRawDataLogExists(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, logger *logrus.Entry) *Error {
	rawDataLogExists, microserviceID, err := s.rawdatalogRepo.Exists(msK8sInfo.Namespace, ms.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to check if Raw Data Log exists")
		return newInternalError(fmt.Errorf("failed to check if Raw Data Log exists: %w", err))
	}
	if !rawDataLogExists {
		logger.Debug("Raw Data Log does not exist, creating a new one")
		return s.createRawDataLog(msK8sInfo, ms, logger)
	} else {
		return s.updateRawDataLogWebhooks(msK8sInfo, ms.Extra.Webhooks, ms.Environment, microserviceID, logger)
	}
}

func (s *Handler) ensurePurchaseOrderAPIExists(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, tenant platform.TenantId, logger *logrus.Entry) *Error {
	exists, err := s.repo.Exists(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		logger.WithError(err).Error("Failed to check if Purchase Order API exists with K8sRepo")
		return newInternalError(fmt.Errorf("failed to check if Purchase Order API exists with K8sRepo: %w", err))
	}
	if !exists {
		logger.WithField("microserviceID", ms.Dolittle.MicroserviceID).Warnf("A Purchase Order API Microservice with the name %s does not exist in K8sRepo", ms.Name)
		return newConflict(fmt.Errorf("a Purchase Order API Microservice with ID %s does not exist in %s environment in application %s under customer %s", ms.Dolittle.MicroserviceID, ms.Environment, ms.Dolittle.ApplicationID, ms.Dolittle.TenantID))
	}
	return nil
}
func (s *Handler) updatePurchaseOrderAPIWebhooks(msK8sInfo k8s.MicroserviceK8sInfo, webhooks []platform.RawDataLogIngestorWebhookConfig, environment, microserviceID string, logger *logrus.Entry) *Error {
	var storedMicroservice platform.HttpInputPurchaseOrderInfo
	bytes, err := s.gitRepo.GetMicroservice(msK8sInfo.Tenant.ID, msK8sInfo.Application.ID, environment, microserviceID)
	if err != nil {
		logger.WithError(err).Error("Failed to get Purchase Order API microservice from GitRepo")
		return newInternalError(fmt.Errorf("failed to get Purchase Order API microservice from GitRepo: %w", err))
	}

	json.Unmarshal(bytes, &storedMicroservice)
	storedMicroservice.Extra.Webhooks = webhooks

	if err := s.gitRepo.SaveMicroservice(storedMicroservice.Dolittle.TenantID, storedMicroservice.Dolittle.ApplicationID, storedMicroservice.Environment, storedMicroservice.Dolittle.MicroserviceID, storedMicroservice); err != nil {
		logger.WithError(err).Error("Failed to save Purchase Order API in GitRepo")
		return newInternalError(fmt.Errorf("failed to save Purchase Order API in GitRepo: %w", err))
	}
	return nil
}

func (s *Handler) createRawDataLog(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, logger *logrus.Entry) *Error {
	rawDataLogMicroservice := s.extractRawDataLogInfo(ms)

	if err := s.rawdatalogRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, rawDataLogMicroservice); err != nil {
		logger.WithError(err).Error("Failed to create Raw Data Log")
		return newInternalError(fmt.Errorf("failed to create Raw Data Log: %w", err))
	}

	if err := s.gitRepo.SaveMicroservice(rawDataLogMicroservice.Dolittle.TenantID, rawDataLogMicroservice.Dolittle.ApplicationID, rawDataLogMicroservice.Environment, rawDataLogMicroservice.Dolittle.MicroserviceID, rawDataLogMicroservice); err != nil {
		logger.WithError(err).Error("Failed to save Raw Data Log in GitRepo")
		return newInternalError(fmt.Errorf("failed to save Raw Data Log in GitRepo: %w", err))
	}
	return nil
}

func (s *Handler) updateRawDataLogWebhooks(msK8sInfo k8s.MicroserviceK8sInfo, webhooks []platform.RawDataLogIngestorWebhookConfig, environment, microserviceID string, logger *logrus.Entry) *Error {
	var storedMicroservice platform.HttpInputRawDataLogIngestorInfo
	bytes, err := s.gitRepo.GetMicroservice(msK8sInfo.Tenant.ID, msK8sInfo.Application.ID, environment, microserviceID)
	if err != nil {
		logger.WithError(err).Error("Failed to get Raw Data Log microservice from GitRepo")
		return newInternalError(fmt.Errorf("failed to get Raw Data Log microservice from GitRepo: %w", err))
	}

	json.Unmarshal(bytes, &storedMicroservice)
	storedMicroservice.Extra.Webhooks = webhooks
	if err := s.rawdatalogRepo.Update(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, storedMicroservice); err != nil {
		logger.WithError(err).Error("Failed to update Raw Data Log")
		return newInternalError(fmt.Errorf("failed to update Raw Data Log: %w", err))
	}

	if err := s.gitRepo.SaveMicroservice(storedMicroservice.Dolittle.TenantID, storedMicroservice.Dolittle.ApplicationID, storedMicroservice.Environment, storedMicroservice.Dolittle.MicroserviceID, storedMicroservice); err != nil {
		logger.WithError(err).Error("Failed to save Raw Data Log in GitRepo")
		return newInternalError(fmt.Errorf("failed to save Raw Data Log in GitRepo: %w", err))
	}
	return nil
}

func (s *Handler) extractRawDataLogInfo(ms platform.HttpInputPurchaseOrderInfo) platform.HttpInputRawDataLogIngestorInfo {
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
