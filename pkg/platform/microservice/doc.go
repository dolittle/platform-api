package microservice

import (
	"fmt"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	TodoCustomersTenantID string = "17426336-fb8e-4425-8ab7-07d488367be9"
)

// RequestHandler defines a system that can handle HTTP requests for creating and deleting microservices
type RequestHandler interface {
	// CanHandle checks whether it can handle the request.
	CanHandle(kind platform.MicroserviceKind, data []byte) bool
	// Create handles the creation of a microservice
	Create(responseWriter http.ResponseWriter, request *http.Request, data []byte, applicationInfo platform.Application) error
	// Delete handles the deletion of a microservice
	Delete(namespace string, microserviceID string) error
}

// Defines a parser that can parse the HTTP request input data to a microservice
type Parser interface {
	// Parses the bytes of an HTTP request and stores the result in the value pointed to by microservice.
	Parse(requestBytes []byte, microservice platform.Microservice, applicationInfo platform.Application) (microserviceK8sInfo, *errors.StatusError)
}
type service struct {
	simpleRepo                 simpleRepo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	rawDataLogIngestorRepo     rawdatalog.RawDataLogIngestorRepo
	purchaseOrderHandler       RequestHandler
	k8sDolittleRepo            platform.K8sRepo
	gitRepo                    storage.Repo
	parser                     Parser
}

type microserviceK8sInfo struct {
	Tenant      k8s.Tenant
	Application k8s.Application
	Namespace   string
}

func createIngress() k8s.Ingress {
	// TODO replace this with something from the cluster or something from git
	domainPrefix := "freshteapot-taco"
	return k8s.Ingress{
		Host:       fmt.Sprintf("%s.dolittle.cloud", domainPrefix),
		SecretName: fmt.Sprintf("%s-certificate", domainPrefix),
	}
}
