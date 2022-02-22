package purchaseorderapi

import (
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	handler         *Handler
	k8sDolittleRepo platformK8s.K8sRepo
	logger          logrus.FieldLogger
}

func NewService(isProduction bool, gitRepo storage.Repo, k8sDolittleRepo platformK8s.K8sRepo, k8sClient kubernetes.Interface, logContext logrus.FieldLogger) service {
	rawDataLogRepo := rawdatalog.NewRawDataLogIngestorRepo(isProduction, k8sDolittleRepo, k8sClient, logContext)
	specFactory := NewK8sResourceSpecFactory()
	k8sResources := NewK8sResource(k8sClient, specFactory)
	k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))
	return service{
		handler: NewHandler(
			parser.NewJsonParser(),
			NewRepo(k8sResources, specFactory, k8sClient, k8sRepoV2),
			gitRepo,
			rawDataLogRepo,
			logContext,
		),
		k8sDolittleRepo: k8sDolittleRepo,
		logger:          logContext,
	}
}

func (s *service) GetDataStatus(responseWriter http.ResponseWriter, request *http.Request) {
	customerID := request.Header.Get("Tenant-ID")

	vars := mux.Vars(request)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]

	logger := s.logger.WithFields(logrus.Fields{
		"service":         "PurchaseOrderAPI",
		"method":          "GetDataStatus",
		"customer_id":     customerID,
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
	})

	dns, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		logger.WithError(err).Error("Failed to get the microservices DNS")
		utils.RespondWithError(responseWriter, http.StatusNotFound, err.Error())
		return
	}

	status, getError := s.handler.GetDataStatus(dns, customerID, applicationID, environment, microserviceID)

	if getError != nil {
		logger.WithError(getError).Error("Failed to get the microservices data status")
		utils.RespondWithError(responseWriter, getError.StatusCode, getError.Error())
		return
	}
	utils.RespondWithJSON(responseWriter, http.StatusAccepted, status)
}
