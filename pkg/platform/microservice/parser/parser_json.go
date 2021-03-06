package parser

import (
	"encoding/json"
	"fmt"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	. "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
)

type parser struct{}

// NewJsonParser creates a new Parser that can parse JSON-encoded data
func NewJsonParser() Parser {
	return &parser{}
}

func (p *parser) Parse(requestBytes []byte, microservice platform.Microservice, applicationInfo platform.Application) (MicroserviceK8sInfo, *errors.StatusError) {
	info := MicroserviceK8sInfo{}

	err := json.Unmarshal(requestBytes, &microservice)
	if err != nil {
		fmt.Println(err)
		return info, errors.NewBadRequest(fmt.Sprintf("Invalid request payload. Error: %v", err))
	}

	info.Customer = k8s.Tenant{
		ID:   applicationInfo.Customer.ID,
		Name: applicationInfo.Customer.Name,
	}

	info.Application = k8s.Application{
		ID:   applicationInfo.ID,
		Name: applicationInfo.Name,
	}
	if info.Customer.ID != microservice.GetBase().Dolittle.CustomerID {
		return info, errors.NewBadRequest("Invalid request payload. Tenant id in the system does not match the one in the input")
	}

	if info.Application.ID != microservice.GetBase().Dolittle.ApplicationID {

		return info, errors.NewBadRequest("Invalid request payload. Currently locked down to application 11b6cf47-5d9f-438f-8116-0d9828654657")
	}

	info.Namespace = fmt.Sprintf("application-%s", info.Application.ID)
	return info, nil
}
