package k8s

import (
	"github.com/dolittle/platform-api/pkg/platform"
	networkingv1 "k8s.io/api/networking/v1"
)

type ShortInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Tenant ShortInfo
type Application ShortInfo

type Microservice struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Tenant      Tenant                    `json:"tenant"`
	Application Application               `json:"application"`
	Environment string                    `json:"environment"`
	Kind        platform.MicroserviceKind `json:"kind"`
}

type SimpleIngressRule struct {
	Path            string                `json:"path"`
	PathType        networkingv1.PathType `json:"path_type"`
	ServiceName     string                `json:"service_name"`
	ServicePortName string                `json:"service_port_name"`
}
