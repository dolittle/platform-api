package microservice

import (
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
)

type service struct {
	simpleRepo                 simple.Repo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	rawDataLogIngestorRepo     rawdatalog.RawDataLogIngestorRepo
	purchaseOrderHandler       *purchaseorderapi.Handler
	k8sDolittleRepo            platformK8s.K8sPlatformRepo
	gitRepo                    storage.Repo
	parser                     parser.Parser
	logContext                 logrus.FieldLogger
}
