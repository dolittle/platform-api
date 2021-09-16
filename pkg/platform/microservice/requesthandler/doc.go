package requesthandler

import (
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
)

// Handler defines a system that can handle HTTP requests for creating and deleting microservices
type Handler interface {
	// Create handles the creation of a microservice
	Create(request *http.Request, input []byte, applicationInfo platform.Application) (platform.Microservice, *Error)
	// Delete handles the deletion of a microservice
	Delete(namespace, microserviceID string) *Error
}

// Defines a parser that can parse the HTTP request input data to a microservice
type Parser interface {
	// Parses the bytes of an HTTP request and stores the result in the value pointed to by microservice.
	Parse(requestBytes []byte, microservice platform.Microservice, applicationInfo platform.Application) (k8s.MicroserviceK8sInfo, *Error)
}
