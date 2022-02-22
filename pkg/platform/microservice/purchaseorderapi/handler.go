package purchaseorderapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle/platform-api/pkg/platform/storage"
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
func (s *Handler) Create(inputBytes []byte, applicationInfo platform.Application, customerTenants []platform.CustomerTenantInfo) (platform.HttpInputPurchaseOrderInfo, *Error) {
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
		"customer_id":    applicationInfo.Customer.ID,
		"application_id": applicationInfo.ID,
		"environment":    ms.Environment,
	})
	logger.Debug("Starting to create a PurchaseOrderAPI microservice")

	exists, statusErr := s.purchaseOrderApiExists(msK8sInfo, ms, logger)
	if statusErr != nil {
		logger.WithError(statusErr).Error("Failed to check whether Purchase Order API exists")
		return ms, newInternalError(fmt.Errorf("failed to whether Purchase Order API exists: %w", statusErr))
	}

	if exists {
		logger.WithField("microserviceID", ms.Dolittle.MicroserviceID).Warn("A Purchase Order API Microservice with the same name already exists in kubernetes or git storage")
		return ms, newConflict(fmt.Errorf("a Purchase Order API Microservice with the same name already exists in kubernetes or git storage"))
	}

	if statusErr := s.ensureRawDataLogExists(msK8sInfo, ms, customerTenants, logger); statusErr != nil {
		return ms, statusErr
	}

	return ms, s.createPurchaseOrderAPI(msK8sInfo, ms, customerTenants, logger)
}

// Update updates an existing PurchaseOrderAPI microservice and creates a RawDataLog microservice too if it didn't already exist
func (s *Handler) UpdateWebhooks(inputBytes []byte, applicationInfo platform.Application, customerTenants []platform.CustomerTenantInfo) (platform.HttpInputPurchaseOrderInfo, *Error) {
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
		"customer_id":    applicationInfo.Customer.ID,
		"application_id": applicationInfo.ID,
		"environment":    ms.Environment,
	})
	logger.Debug("Starting to update PurchaseOrderAPI microservice")

	exists, statusErr := s.purchaseOrderApiExists(msK8sInfo, ms, logger)
	if statusErr != nil {
		logger.WithError(statusErr).Error("Failed to check whether Purchase Order API exists")
		return ms, newInternalError(fmt.Errorf("failed to whether Purchase Order API exists: %w", statusErr))
	}
	if !exists {
		logger.WithField("microserviceID", ms.Dolittle.MicroserviceID).Warn("A Purchase Order API Microservice does not exist in kubernetes or git storage")
		return ms, newConflict(fmt.Errorf("a Purchase Order API Microservice does not exist in kubernetes or git storage"))
	}

	if statusErr := s.ensureRawDataLogExists(msK8sInfo, ms, customerTenants, logger); statusErr != nil {
		return ms, statusErr
	}
	return ms, s.updatePurchaseOrderAPIWebhooks(msK8sInfo, ms.Extra.Webhooks, ms.Environment, ms.Dolittle.MicroserviceID, logger)
}

func (s *Handler) Delete(applicationID, environment, microserviceID string) error {
	if err := s.repo.Delete(applicationID, environment, microserviceID); err != nil {
		return fmt.Errorf("failed to delete Purchase Order API: %w", err)
	}
	return nil
}

func (s *Handler) GetDataStatus(dns, customerID, applicationID, environment, microserviceID string) (platform.PurchaseOrderStatus, *Error) {
	logger := s.logContext.WithFields(logrus.Fields{
		"handler":         "PurchaseOrderAPI",
		"method":          "GetDataStatus",
		"tenant_id":       customerID,
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
		"dns":             dns,
	})

	var status platform.PurchaseOrderStatus
	url := fmt.Sprintf("http://%s%s", dns, "/api/purchaseorders/datastatus")
	resp, err := http.Get(url)

	if err != nil {
		logger.WithError(err).Error("Failed to request Purchase Order API microservices status")
		return status, newInternalError(fmt.Errorf("failed to get Purchase Order API microservices data status"))
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.New(resp.Status)
		error := &Error{StatusCode: resp.StatusCode, Err: err}
		logger.WithError(err).Error("Purchase Order API microservice data status didn't return 2** status")
		return status, error
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithError(err).Error("Failed to read response for Purchase Order API microservices data status")
		return status, newInternalError(fmt.Errorf("failed to read response for Purchase Order API microservices data status"))
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		logger.WithError(err).Error("Failed to parse response for Purchase Order API microservices status")
		return status, newInternalError(fmt.Errorf("failed to parse response for Purchase Order API microservices status"))
	}
	return status, nil
}

func (s *Handler) purchaseOrderApiExists(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, logger *logrus.Entry) (bool, *Error) {
	microservices, err := s.gitRepo.GetMicroservices(msK8sInfo.Customer.ID, msK8sInfo.Application.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to get microservices from GitRepo")
		return false, newInternalError(fmt.Errorf("failed to get microservices from GitRepo: %w", err))
	}
	for _, microservice := range microservices {
		if microservice.Kind == platform.MicroserviceKindPurchaseOrderAPI && strings.EqualFold(microservice.Environment, ms.Environment) {
			return true, nil
		}
	}

	exists, err := s.repo.EnvironmentHasPurchaseOrderAPI(msK8sInfo.Namespace, ms)
	if err != nil {
		logger.WithError(err).Error("Failed to check if environment has Purchase Order API with K8sRepo")
		return false, newInternalError(fmt.Errorf("failed to check if environment has Purchase Order API with K8sRepo: %w", err))
	}
	if exists {
		return true, nil
	}
	return false, nil
}

func (s *Handler) createPurchaseOrderAPI(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, customerTenants []platform.CustomerTenantInfo, logger *logrus.Entry) *Error {
	if err := s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Customer, msK8sInfo.Application, customerTenants, ms); err != nil {
		logger.WithError(err).Error("Failed to create Purchase Order API")
		return newInternalError(fmt.Errorf("failed to create Purchase Order API: %w", err))
	}

	if err := s.gitRepo.SaveMicroservice(ms.Dolittle.CustomerID, ms.Dolittle.ApplicationID, ms.Environment, ms.Dolittle.MicroserviceID, ms); err != nil {
		// TODO change
		logger.WithError(err).Error("Failed to save Purchase Order API in GitRepo")
		return newInternalError(fmt.Errorf("failed to save Purchase Order API in GitRepo"))
	}
	return nil
}

func (s *Handler) ensureRawDataLogExists(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, customerTenants []platform.CustomerTenantInfo, logger *logrus.Entry) *Error {
	rawDataLogExists, microserviceID, err := s.rawdatalogRepo.Exists(msK8sInfo.Namespace, ms.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to check if Raw Data Log exists")
		return newInternalError(fmt.Errorf("failed to check if Raw Data Log exists: %w", err))
	}
	if !rawDataLogExists {
		logger.Debug("Raw Data Log does not exist, creating a new one")
		return s.createRawDataLog(msK8sInfo, ms, customerTenants, logger)
	} else {
		return s.updateRawDataLogWebhooks(msK8sInfo, ms.Extra.Webhooks, ms.Environment, microserviceID, logger)
	}
}

func (s *Handler) updatePurchaseOrderAPIWebhooks(msK8sInfo k8s.MicroserviceK8sInfo, webhooks []platform.RawDataLogIngestorWebhookConfig, environment, microserviceID string, logger *logrus.Entry) *Error {
	var storedMicroservice platform.HttpInputPurchaseOrderInfo
	bytes, err := s.gitRepo.GetMicroservice(msK8sInfo.Customer.ID, msK8sInfo.Application.ID, environment, microserviceID)
	if err != nil {
		logger.WithError(err).Error("Failed to get Purchase Order API microservice from GitRepo")
		return newInternalError(fmt.Errorf("failed to get Purchase Order API microservice from GitRepo: %w", err))
	}

	json.Unmarshal(bytes, &storedMicroservice)
	storedMicroservice.Extra.Webhooks = webhooks

	if err := s.gitRepo.SaveMicroservice(storedMicroservice.Dolittle.CustomerID, storedMicroservice.Dolittle.ApplicationID, storedMicroservice.Environment, storedMicroservice.Dolittle.MicroserviceID, storedMicroservice); err != nil {
		logger.WithError(err).Error("Failed to save Purchase Order API in GitRepo")
		return newInternalError(fmt.Errorf("failed to save Purchase Order API in GitRepo: %w", err))
	}
	return nil
}

func (s *Handler) createRawDataLog(msK8sInfo k8s.MicroserviceK8sInfo, ms platform.HttpInputPurchaseOrderInfo, customerTenants []platform.CustomerTenantInfo, logger *logrus.Entry) *Error {
	rawDataLogMicroservice := s.extractRawDataLogInfo(ms)
	if err := s.rawdatalogRepo.Create(msK8sInfo.Namespace, msK8sInfo.Customer, msK8sInfo.Application, customerTenants, rawDataLogMicroservice); err != nil {
		logger.WithError(err).Error("Failed to create Raw Data Log")
		return newInternalError(fmt.Errorf("failed to create Raw Data Log: %w", err))
	}

	if err := s.gitRepo.SaveMicroservice(rawDataLogMicroservice.Dolittle.CustomerID, rawDataLogMicroservice.Dolittle.ApplicationID, rawDataLogMicroservice.Environment, rawDataLogMicroservice.Dolittle.MicroserviceID, rawDataLogMicroservice); err != nil {
		logger.WithError(err).Error("Failed to save Raw Data Log in GitRepo")
		return newInternalError(fmt.Errorf("failed to save Raw Data Log in GitRepo: %w", err))
	}
	return nil
}

func (s *Handler) updateRawDataLogWebhooks(msK8sInfo k8s.MicroserviceK8sInfo, webhooks []platform.RawDataLogIngestorWebhookConfig, environment, microserviceID string, logger *logrus.Entry) *Error {
	var storedMicroservice platform.HttpInputRawDataLogIngestorInfo
	bytes, err := s.gitRepo.GetMicroservice(msK8sInfo.Customer.ID, msK8sInfo.Application.ID, environment, microserviceID)
	if err != nil {
		logger.WithError(err).Error("Failed to get Raw Data Log microservice from GitRepo")
		return newInternalError(fmt.Errorf("failed to get Raw Data Log microservice from GitRepo: %w", err))
	}

	json.Unmarshal(bytes, &storedMicroservice)
	storedMicroservice.Extra.Webhooks = webhooks
	if err := s.rawdatalogRepo.Update(msK8sInfo.Namespace, msK8sInfo.Customer, msK8sInfo.Application, storedMicroservice); err != nil {
		logger.WithError(err).Error("Failed to update Raw Data Log")
		return newInternalError(fmt.Errorf("failed to update Raw Data Log: %w", err))
	}

	if err := s.gitRepo.SaveMicroservice(storedMicroservice.Dolittle.CustomerID, storedMicroservice.Dolittle.ApplicationID, storedMicroservice.Environment, storedMicroservice.Dolittle.MicroserviceID, storedMicroservice); err != nil {
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
				CustomerID:     ms.Dolittle.CustomerID,
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
