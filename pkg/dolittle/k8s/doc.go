package k8s

import (
	"github.com/dolittle/platform-api/pkg/platform"
	networkingv1 "k8s.io/api/networking/v1"
)

type Tenant struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Application struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Ingress struct {
	Host       string `json:"host"`
	SecretName string `json:"secret_name"`
}

type Microservice struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Tenant      Tenant      `json:"tenant"`
	Application Application `json:"application"`
	Environment string      `json:"environment"`
	// Linked to TenantsCustomer (look at ingress.go for now)
	ResourceID string                    `json:"resource_id"`
	Kind       platform.MicroserviceKind `json:"kind"`
}

type SimpleIngressRule struct {
	Path            string                `json:"path"`
	PathType        networkingv1.PathType `json:"path_type"`
	ServiceName     string                `json:"service_name"`
	ServicePortName string                `json:"service_port_name"`
}
