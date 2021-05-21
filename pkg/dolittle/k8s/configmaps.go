package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type emptyObject struct{}
type MicroserviceResources map[string]MicroserviceResource

type MicroserviceResource struct {
	Readmodels MicroserviceResourceReadmodels `json:"readModels"`
	Eventstore MicroserviceResourceEventstore `json:"eventStore"`
}
type MicroserviceResourceReadmodels struct {
	Host     string `json:"host"`
	Database string `json:"database"`
	Usessl   bool   `json:"useSSL"`
}
type MicroserviceResourceEventstore struct {
	Servers  []string `json:"servers"`
	Database string   `json:"database"`
}

type MicroserviceEndpoints struct {
	Public  MicroserviceEndpointPort `json:"public"`
	Private MicroserviceEndpointPort `json:"private"`
}
type MicroserviceEndpointPort struct {
	Port int `json:"port"`
}

type Appsettings struct {
	Logging AppsettingsLogging `json:"Logging"`
}
type AppsettingsLoglevel struct {
	Default   string `json:"Default"`
	System    string `json:"System"`
	Microsoft string `json:"Microsoft"`
}
type AppsettingsConsole struct {
	Includescopes   bool   `json:"IncludeScopes"`
	Timestampformat string `json:"TimestampFormat"`
}
type AppsettingsLogging struct {
	Includescopes bool                `json:"IncludeScopes"`
	Loglevel      AppsettingsLoglevel `json:"LogLevel"`
	Console       AppsettingsConsole  `json:"Console"`
}

func NewConfigFilesConfigmap(microservice Microservice) *corev1.ConfigMap {
	name := fmt.Sprintf("%s-%s-config-files",
		microservice.Environment,
		microservice.Name,
	)

	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)

	name = strings.ToLower(name)

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
	}
}

func NewEnvVariablesConfigmap(microservice Microservice) *corev1.ConfigMap {
	name := fmt.Sprintf("%s-%s-env-variables",
		microservice.Environment,
		microservice.Name,
	)

	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)

	name = strings.ToLower(name)

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
	}
}

func NewMicroserviceConfigmap(microservice Microservice, customersTenantID string) *corev1.ConfigMap {
	name := fmt.Sprintf("%s-%s-dolittle",
		microservice.Environment,
		microservice.Name,
	)

	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)

	databasePrefix := fmt.Sprintf("%s_%s_%s",
		microservice.Application.Name,
		microservice.Environment,
		microservice.Name,
	)

	name = strings.ToLower(name)
	databasePrefix = strings.ToLower(databasePrefix)

	resources := MicroserviceResources{
		customersTenantID: MicroserviceResource{
			Readmodels: MicroserviceResourceReadmodels{
				Host:     fmt.Sprintf("mongodb://dev-mongo.application-%s.svc.cluster.local:27017", microservice.Application.ID),
				Database: fmt.Sprintf("%s_readmodels", databasePrefix),
				Usessl:   false,
			},
			Eventstore: MicroserviceResourceEventstore{
				Servers: []string{
					fmt.Sprintf("dev-mongo.application-%s.svc.cluster.local", microservice.Application.ID),
				},
				Database: fmt.Sprintf("%s_eventstore", databasePrefix),
			},
		},
	}

	endpoints := MicroserviceEndpoints{
		Public: MicroserviceEndpointPort{
			Port: 50052,
		},
		Private: MicroserviceEndpointPort{
			Port: 50053,
		},
	}

	appsettings := Appsettings{
		Logging: AppsettingsLogging{
			Includescopes: false,
			Loglevel: AppsettingsLoglevel{
				Default:   "Debug",
				System:    "Information",
				Microsoft: "Information",
			},
			Console: AppsettingsConsole{
				Includescopes:   true,
				Timestampformat: "[yyyy-MM-dd HH:mm:ss] ",
			},
		},
	}
	b, _ := json.MarshalIndent(appsettings, "", "  ")
	appsettingsJSON := string(b)

	b, _ = json.MarshalIndent(endpoints, "", "  ")
	endpointsJSON := string(b)

	b, _ = json.MarshalIndent(resources, "", "  ")
	resourcesJSON := string(b)

	b, _ = json.MarshalIndent(emptyObject{}, "", "  ")
	eventHorizonsJSON := string(b)

	b, _ = json.MarshalIndent(emptyObject{}, "", "  ")
	eventHorizonConsentsJSON := string(b)

	b, _ = json.MarshalIndent(emptyObject{}, "", "  ")
	microservicesJSON := string(b)

	// TODO the json objects are ugly :)
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
		Data: map[string]string{
			"resources.json":              resourcesJSON,
			"event-horizons.json":         eventHorizonsJSON,
			"event-horizon-consents.json": eventHorizonConsentsJSON,
			"microservices.json":          microservicesJSON,
			"endpoints.json":              endpointsJSON,
			"appsettings.json":            appsettingsJSON,
		},
	}
}
