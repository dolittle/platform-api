package requesthandler

import (
	_ "embed"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type purchaseOrderApiHandler struct {
	parser         Parser
	repo           purchaseorderapi.Repo
	gitRepo        storage.Repo
	rawdatalogRepo rawdatalog.RawDataLogIngestorRepo
	logContext     logrus.FieldLogger
}

func NewPurchaseOrderApiHandler(parser Parser, repo purchaseorderapi.Repo, gitRepo storage.Repo, rawDataLogIngestorRepo rawdatalog.RawDataLogIngestorRepo, logContext logrus.FieldLogger) Handler {
	return &purchaseOrderApiHandler{parser, repo, gitRepo, rawDataLogIngestorRepo, logContext}
}

func (s *purchaseOrderApiHandler) CanHandle(kind platform.MicroserviceKind) bool {
	return kind == platform.MicroserviceKindPurchaseOrderAPI
}

// Create creates a new PurchaseOrderAPI microservice and creates a RawDataLog microservice too if it didn't already exist
func (s *purchaseOrderApiHandler) Create(request *http.Request, inputBytes []byte, applicationInfo platform.Application) (platform.Microservice, *Error) {
	// Function assumes access check has taken place
	var ms platform.HttpInputPurchaseOrderInfo
	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		return nil, parserError
	}
	s.logContext.WithFields(logrus.Fields{
		"method":        "Create",
		"tenantID":      applicationInfo.Tenant.ID,
		"applicationID": applicationInfo.ID,
		"environment":   ms.Environment,
	}).Debug("Starting to create a PurchaseOrderAPI microservice")

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
		ms)
	if err != nil {
		return nil, NewInternalError(err)
	}
	return ms, s.ensureRawDataLogIngestorExists(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ms)
}

func (s *purchaseOrderApiHandler) Delete(namespace string, microserviceID string) *Error {
	if err := s.repo.Delete(namespace, microserviceID); err != nil {
		return NewInternalError(err)
	}
	return nil
}

func (s *purchaseOrderApiHandler) ensureRawDataLogIngestorExists(namespace string, customer k8s.Tenant, application k8s.Application, purchaseOrderApi platform.HttpInputPurchaseOrderInfo) *Error {

	rawDataLogExists, err := s.rawdatalogRepo.Exists(namespace, purchaseOrderApi.Environment)
	if err != nil {
		return NewInternalError(err)
	}
	if rawDataLogExists {
		return nil
	}
	storedIngress, err := getFirstIngress(s.gitRepo, purchaseOrderApi.Dolittle.TenantID, purchaseOrderApi.Dolittle.ApplicationID, purchaseOrderApi.Environment)
	if err != nil {
		return NewInternalError(err)
	}

	input := s.createRawDataLogIngestorInfo(purchaseOrderApi, storedIngress)
	applicationIngress := k8s.Ingress{
		Host:       storedIngress.Host,
		SecretName: storedIngress.SecretName,
	}
	err = s.rawdatalogRepo.Create(namespace, customer, application, applicationIngress, input)

	if err != nil {
		return NewInternalError(err)
	}

	err = s.gitRepo.SaveMicroservice(
		input.Dolittle.TenantID,
		input.Dolittle.ApplicationID,
		input.Environment,
		input.Dolittle.MicroserviceID,
		input)
	if err != nil {
		return NewInternalError(err)
	}
	return nil
}
func (s *purchaseOrderApiHandler) createRawDataLogIngestorInfo(purchaseOrderApi platform.HttpInputPurchaseOrderInfo, storedIngress platform.EnvironmentIngress) platform.HttpInputRawDataLogIngestorInfo {
	webhookConfigs := []platform.RawDataLogIngestorWebhookConfig{}

	for _, webhook := range purchaseOrderApi.Extra.Webhooks {
		webhook := platform.RawDataLogIngestorWebhookConfig{
			Kind:          webhook.Kind,
			UriSuffix:     webhook.UriSuffix,
			Authorization: webhook.Authorization,
		}
		webhookConfigs = append(webhookConfigs, webhook)
	}

	return platform.HttpInputRawDataLogIngestorInfo{
		MicroserviceBase: platform.MicroserviceBase{
			Name:        purchaseOrderApi.Extra.RawDataLogName,
			Kind:        platform.MicroserviceKindRawDataLogIngestor,
			Environment: purchaseOrderApi.Environment,
			Dolittle: platform.HttpInputDolittle{
				ApplicationID:  purchaseOrderApi.Dolittle.ApplicationID,
				TenantID:       purchaseOrderApi.Dolittle.TenantID,
				MicroserviceID: uuid.New().String(),
			},
		},
		Extra: platform.HttpInputRawDataLogIngestorExtra{
			// TODO these images won't evolve automatically
			Headimage:    "dolittle/platform-api:latest",
			Runtimeimage: "dolittle/runtime:6.1.0",
			Ingress: platform.HttpInputSimpleIngress{
				Host:             storedIngress.Host,
				DomainPrefix:     storedIngress.DomainPrefix,
				SecretNamePrefix: storedIngress.SecretName,
				// TODO this is now hardcoded
				Pathtype: "Prefix",
				Path:     "/api/webhooks",
			},
			WriteTo:  "nats",
			Webhooks: webhookConfigs,
		},
	}
}
