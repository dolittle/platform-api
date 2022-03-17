package microservice

import (
	"github.com/dolittle/platform-api/pkg/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
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
		logContext: logContext,
	}
}
