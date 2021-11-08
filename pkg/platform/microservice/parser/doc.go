package parser

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Defines a parser that can parse the HTTP request input data to a microservice
type Parser interface {
	// Parses the bytes of an HTTP request and stores the result in the value pointed to by microservice.
	Parse(requestBytes []byte, microservice platform.Microservice, applicationInfo platform.Application) (k8s.MicroserviceK8sInfo, *errors.StatusError)
}
