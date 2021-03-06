package automate

import (
	"encoding/json"

	"github.com/dolittle/platform-api/pkg/platform"
)

// ParseMicroserviceMetadata Given a json oneliner, return the microservice metadata
func ParseMicroserviceMetadata(input string) (platform.MicroserviceMetadataShortInfo, error) {
	var data platform.MicroserviceMetadataShortInfo
	err := json.Unmarshal([]byte(input), &data)
	return data, err
}
