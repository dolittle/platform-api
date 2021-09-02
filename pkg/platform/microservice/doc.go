package microservice

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
)

type service struct {
	simpleRepo                 simpleRepo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	rawDataLogIngestorRepo     rawdatalog.RawDataLogIngestorRepo
	purchaseOrderAPIRepo       purchaseorderapi.PurchaseOrderAPIRepo
	k8sDolittleRepo            platform.K8sRepo
	gitRepo                    storage.Repo
}
