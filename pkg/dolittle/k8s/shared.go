package k8s

import (
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
)

func GetLabels(microservice Microservice) map[string]string {
	return platformK8s.GetLabelsForMicroservice(
		microservice.Tenant.Name,
		microservice.Application.Name,
		microservice.Environment,
		microservice.Name,
	)
}

func GetAnnotations(microservice Microservice) map[string]string {
	return platformK8s.GetAnnotationsForMicroservice(
		microservice.Tenant.ID,
		microservice.Application.ID,
		microservice.ID,
		string(microservice.Kind),
	)
}
