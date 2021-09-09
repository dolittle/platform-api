package requests

import (
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
)

// RequestHandler defines a system that can handle HTTP requests for creating and deleting microservices
type RequestHandler interface {
	// Create handles the creation of a microservice
	Create(responseWriter http.ResponseWriter, request *http.Request, data []byte, applicationInfo platform.Application) error
	// Delete handles the deletion of a microservice
	Delete(namespace, microserviceID string) error
}
