package k8s

import "strings"

func ParseLabel(input string) string {
	// https://app.asana.com/0/0/1201457681486811/f
	// a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')"
	// I wonder if we can invert and replace each match with "_"
	// (([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')
	// https://github.com/kubernetes/apimachinery/blob/master/pkg/util/validation/validation.go
	cleaned := strings.ReplaceAll(input, " ", "_")
	return cleaned
}

func GetLabelsForCustomer(tenant string) map[string]string {
	return map[string]string{
		"tenant": ParseLabel(tenant),
	}
}

func GetAnnotationsForCustomer(customerID string) map[string]string {
	return map[string]string{
		"dolittle.io/tenant-id": ParseLabel(customerID),
	}
}

func GetLabelsForApplication(tenant string, application string) map[string]string {
	return map[string]string{
		"tenant":      ParseLabel(tenant),
		"application": ParseLabel(application),
	}
}

func GetAnnotationsForApplication(customerID string, applicationID string) map[string]string {
	return map[string]string{
		"dolittle.io/tenant-id":      ParseLabel(customerID),
		"dolittle.io/application-id": ParseLabel(applicationID),
	}
}

func GetLabelsForEnvironment(tenant string, application string, environment string) map[string]string {
	return map[string]string{
		"tenant":      ParseLabel(tenant),
		"application": ParseLabel(application),
		"environment": ParseLabel(environment),
	}
}

func GetLabelsForMicroservice(tenant string, application string, environment string, name string) map[string]string {
	return map[string]string{
		"tenant":       ParseLabel(tenant),
		"application":  ParseLabel(application),
		"environment":  ParseLabel(environment),
		"microservice": ParseLabel(name),
	}
}

func GetAnnotationsForMicroservice(customerID string, applicationID string, microserviceID string, microserviceKind string) map[string]string {
	return map[string]string{
		"dolittle.io/tenant-id":         ParseLabel(customerID),
		"dolittle.io/application-id":    ParseLabel(applicationID),
		"dolittle.io/microservice-id":   ParseLabel(microserviceID),
		"dolittle.io/microservice-kind": ParseLabel(microserviceKind),
	}
}
