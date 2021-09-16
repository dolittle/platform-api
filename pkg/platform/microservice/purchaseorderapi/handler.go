package purchaseorderapi

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/requests"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
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

	var ms platform.HttpInputPurchaseOrderInfo
	msK8sInfo, parserError := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if parserError != nil {
		utils.RespondWithStatusError(responseWriter, parserError)
		return parserError
	}

	s.logContext.WithFields(logrus.Fields{
		"method":        "Create",
		"tenantID":      applicationInfo.Tenant.ID,
		"applicationID": applicationInfo.ID,
		"environment":   ms.Environment,
	}).Debug("Starting to create a PurchaseOrderAPI microservice")

	application, err := s.gitRepo.GetApplication(applicationInfo.Tenant.ID, applicationInfo.ID)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}

	tenant, err := application.GetTenantForEnvironment(ms.Environment)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}

	err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return nil
	}

	rawDataLogExists, err := s.rawdatalogRepo.Exists(msK8sInfo.Namespace, ms.Environment)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}
	if !rawDataLogExists {
		storedIngress := platform.EnvironmentIngress{}
		application, err := s.gitRepo.GetApplication(ms.Dolittle.TenantID, ms.Dolittle.ApplicationID)
		if err != nil {
			return err
		}
		tenant, err := application.GetTenantForEnvironment(ms.Environment)
		if err != nil {
			return err
		}
		storedIngress, ok := application.Environments[funk.IndexOf(application.Environments, func(e platform.HttpInputEnvironment) bool {
			return e.Name == ms.Environment
		})].Ingresses[tenant]

		if !ok {
			return fmt.Errorf("Failed to get stored ingress for tenant %s in environment %s", string(tenant), ms.Environment)
		}

		webhookConfigs := []platform.RawDataLogIngestorWebhookConfig{}

		for _, webhook := range ms.Extra.Webhooks {
			webhook := platform.RawDataLogIngestorWebhookConfig{
				Kind:          webhook.Kind,
				UriSuffix:     webhook.UriSuffix,
				Authorization: webhook.Authorization,
			}
			webhookConfigs = append(webhookConfigs, webhook)
		}

		input := platform.HttpInputRawDataLogIngestorInfo{
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
		applicationIngress := k8s.Ingress{
			Host:       storedIngress.Host,
			SecretName: storedIngress.SecretName,
		}
		s.rawdatalogRepo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, applicationIngress, input)
	}

	err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, tenant, ms)
	if err != nil {
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
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return err
	}

	utils.RespondWithJSON(responseWriter, http.StatusOK, ms)
	return nil
}

func (s *RequestHandler) Delete(namespace string, microserviceID string) error {
	return s.repo.Delete(namespace, microserviceID)

}
