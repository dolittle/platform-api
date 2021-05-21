package k8s

func GetLabels(microservice Microservice) map[string]string {
	return map[string]string{
		"tenant":       microservice.Tenant.Name,
		"application":  microservice.Application.Name,
		"environment":  microservice.Environment,
		"microservice": microservice.Name,
	}
}

func GetAnnotations(microservice Microservice) map[string]string {
	return map[string]string{
		"dolittle.io/tenant-id":       microservice.Tenant.ID,
		"dolittle.io/application-id":  microservice.Application.ID,
		"dolittle.io/microservice-id": microservice.ID,
	}
}
