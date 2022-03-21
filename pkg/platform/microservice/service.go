package microservice

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	simpleRepo                 simple.Repo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	rawDataLogIngestorRepo     rawdatalog.RawDataLogIngestorRepo
	purchaseOrderHandler       *purchaseorderapi.Handler
	k8sDolittleRepo            platformK8s.K8sRepo
	gitRepo                    storage.Repo
	parser                     parser.Parser
	logContext                 logrus.FieldLogger
}

func NewService(
	isProduction bool,
	gitRepo storage.Repo,
	k8sDolittleRepo platformK8s.K8sRepo,
	k8sClient kubernetes.Interface,
	simpleRepo simple.Repo,
	logContext logrus.FieldLogger,
) service {
	parser := parser.NewJsonParser()
	rawDataLogRepo := rawdatalog.NewRawDataLogIngestorRepo(isProduction, k8sDolittleRepo, k8sClient, logContext)
	specFactory := purchaseorderapi.NewK8sResourceSpecFactory()
	k8sResources := purchaseorderapi.NewK8sResource(k8sClient, specFactory)
	k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

	return service{
		gitRepo:                    gitRepo,
		simpleRepo:                 simpleRepo,
		businessMomentsAdaptorRepo: NewBusinessMomentsAdaptorRepo(k8sClient, isProduction),
		rawDataLogIngestorRepo:     rawDataLogRepo,
		k8sDolittleRepo:            k8sDolittleRepo,
		parser:                     parser,
		purchaseOrderHandler: purchaseorderapi.NewHandler(
			parser,
			purchaseorderapi.NewRepo(k8sResources, specFactory, k8sClient, k8sRepoV2),
			gitRepo,
			rawDataLogRepo,
			logContext),
		logContext: logContext.WithFields(logrus.Fields{
			"service": "microservice",
		}),
	}
}

var NotAllowed = errors.New("not allowed to modify")
var EnvironmentDoesNotExist = errors.New("the environment doesn't exist")
var AutomationDisabled = errors.New("automation is disabled for the environment")
var ApplicationNotFound = errors.New("application not found")
var UnknownMicroserviceKind = errors.New("unknown microservice kind")

func (s *service) Create(customerID, applicationID, userID string, microserviceBase platform.HttpMicroserviceBase) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method": "Create",
	})
	studioInfo, err := storage.GetStudioInfo(s.gitRepo, customerID, applicationID, logContext)

	if err != nil {
		// utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return err
	}

	// Confirm application exists
	storedApplication, err := s.gitRepo.GetApplication(customerID, applicationID)
	if err != nil {
		// utils.RespondWithError(w, http.StatusBadRequest, "Not able to find application in the storage")
		return err
	}

	// TODO Confirm the application exists
	// storedApplication.Status.State == storage.BuildStatusStateFinishedSuccess
	// - is it in terraform?
	// - is it being made

	// Confirm user has access to this application + customer
	allowed, err := s.k8sDolittleRepo.CanModifyApplication(customerID, applicationID, userID)
	if !allowed {
		// @joel some type of error here
		return NotAllowed
	}

	environment := microserviceBase.Environment
	exists := storage.EnvironmentExists(storedApplication.Environments, environment)

	if !exists {
		// utils.RespondWithError(w, http.StatusBadRequest, "Unable to add a microservice to an environment that does not exist")
		// @joel some type of error here
		return EnvironmentDoesNotExist
	}

	// Check if automation enabled
	if !s.gitRepo.IsAutomationEnabledWithStudioConfig(studioInfo.StudioConfig, applicationID, environment) {
		// utils.RespondWithError(
		// 	w,
		// 	http.StatusBadRequest,
		// 	fmt.Sprintf(
		// 		"Customer %s with application %s in environment %s does not allow changes via Studio",
		// 		customer.ID,
		// 		applicationID,
		// 		environment,
		// 	),
		// )
		// @joel error about automation being disabled for the environment
		return AutomationDisabled
	}

	customerTenants := storage.GetCustomerTenantsByEnvironment(storedApplication, environment)

	applicationInfo, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			// utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			// @joel k8s error
			return err
		}

		// utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{
		// 	"error": fmt.Sprintf("Application %s not found", applicationID),
		// })
		// @joel application not found error
		return ApplicationNotFound
	}

	// TODO check path
	switch microserviceBase.Kind {
	case platform.MicroserviceKindSimple:
		// @joel move handleSimpleMicroservice to this service
		s.handleSimpleMicroservice(w, request, requestBytes, applicationInfo, customerTenants)
	// TODO let us get simple to work and we can come back to these others.
	//case platform.MicroserviceKindBusinessMomentsAdaptor:
	//	s.handleBusinessMomentsAdaptor(w, request, requestBytes, applicationInfo, customerTenants)
	//case platform.MicroserviceKindRawDataLogIngestor:
	//s.handleRawDataLogIngestor(w, request, requestBytes, applicationInfo, customerTenants)
	//case platform.MicroserviceKindPurchaseOrderAPI:
	//	purchaseOrderAPI, err := s.purchaseOrderHandler.Create(requestBytes, applicationInfo, customerTenants)
	//	if err != nil {
	//		utils.RespondWithError(w, err.StatusCode, err.Error())
	//		break
	//	}
	//	utils.RespondWithJSON(w, http.StatusAccepted, purchaseOrderAPI)
	default:
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Kind %s is not supported", microserviceBase.Kind))
		// @joel handle wrong microservice kind being given
		return UnknownMicroserviceKind
	}

}
