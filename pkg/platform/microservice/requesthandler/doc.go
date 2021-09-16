package requesthandler

import (
	"fmt"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/businessmomentsadaptor"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/thoas/go-funk"
	"k8s.io/client-go/kubernetes"
)

// Handler defines a system that can handle HTTP requests for creating and deleting microservices
type Handler interface {
	// Create handles the creation of a microservice
	Create(request *http.Request, input []byte, applicationInfo platform.Application) (platform.Microservice, *Error)
	// Delete handles the deletion of a microservice
	Delete(namespace, microserviceID string) *Error
	// CanHandle checks whether it can handle this microservice kind
	CanHandle(kind platform.MicroserviceKind) bool
}

// Defines a parser that can parse the HTTP request input data to a microservice
type Parser interface {
	// Parses the bytes of an HTTP request and stores the result in the value pointed to by microservice.
	Parse(requestBytes []byte, microservice platform.Microservice, applicationInfo platform.Application) (k8s.MicroserviceK8sInfo, *Error)
}

type Handlers []Handler

// CreateHandlers creates all the handlers for handling requests for the different microservice kinds
func CreateHandlers(parser Parser, k8sClient kubernetes.Interface, gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo) Handlers {
	rawDataLogRepo := rawdatalog.NewRawDataLogIngestorRepo(k8sDolittleRepo, k8sClient, gitRepo)
	return []Handler{
		NewSimpleHandler(parser, simple.NewSimpleRepo(k8sClient), gitRepo),
		createPurchaseOrderApiHandler(parser, k8sClient, gitRepo, rawDataLogRepo),
		NewRawDataLogIngestorHandler(parser, rawDataLogRepo, gitRepo),
		NewBusinessMomentsAdapterHandler(parser, businessmomentsadaptor.NewBusinessMomentsAdaptorRepo(k8sClient), gitRepo),
	}
}
func createPurchaseOrderApiHandler(parser Parser, k8sClient kubernetes.Interface, gitRepo storage.Repo, rawDataLogRepo rawdatalog.RawDataLogIngestorRepo) Handler {
	specFactory := purchaseorderapi.NewK8sResourceSpecFactory()
	k8sResources := purchaseorderapi.NewK8sResource(k8sClient, specFactory)

	return NewPurchaseOrderApiHandler(
		parser,
		purchaseorderapi.NewRepo(k8sResources, rawDataLogRepo),
		gitRepo)
}

func (h Handlers) GetForKind(kind platform.MicroserviceKind) (Handler, error) {
	foundHandlers := funk.Filter(h, func(foundHandler Handler) bool {
		return foundHandler.CanHandle(kind)
	}).([]Handler)
	if len(foundHandlers) == 0 {
		return nil, fmt.Errorf("No handler that can handle microservice kind %s exists", kind)
	}
	if len(foundHandlers) > 1 {
		return nil, fmt.Errorf("Multiple handlers that can handle microservice kind %s exists", kind)
	}
	return foundHandlers[0], nil
}
