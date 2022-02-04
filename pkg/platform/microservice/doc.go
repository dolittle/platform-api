package microservice

import (
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
)

type service struct {
	simpleRepo                 simpleRepo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	rawDataLogIngestorRepo     rawdatalog.RawDataLogIngestorRepo
	purchaseOrderHandler       *purchaseorderapi.Handler
	k8sDolittleRepo            platformK8s.K8sRepo
	gitRepo                    storage.Repo
	parser                     parser.Parser
	logContext                 logrus.FieldLogger
}
