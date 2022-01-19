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
	Readmodels  MicroserviceResourceReadmodels `json:"readModels"`
	Eventstore  MicroserviceResourceStore      `json:"eventStore"`
	Projections MicroserviceResourceStore      `json:"projections"`
	Embeddings  MicroserviceResourceStore      `json:"embeddings"`
}
type MicroserviceResourceReadmodels struct {
	Host     string `json:"host"`
	Database string `json:"database"`
	Usessl   bool   `json:"useSSL"`
}
type MicroserviceResourceStore struct {
	Servers  []string `json:"servers"`
	Database string   `json:"database"`
}
type MicroserviceEndpoints struct {
	Public     MicroserviceEndpointPort `json:"public"`
	Private    MicroserviceEndpointPort `json:"private"`
	Management MicroserviceEndpointPort `json:"management"`
}
type MicroserviceEndpointPort struct {
	Port int `json:"port"`
}

// platform.json
type MicroservicePlatform struct {
	Applicationname  string `json:"applicationName"`
	Applicationid    string `json:"applicationID"`
	Microservicename string `json:"microserviceName"`
	Microserviceid   string `json:"microserviceID"`
	Customername     string `json:"customerName"`
	Customerid       string `json:"customerID"`
	Environment      string `json:"environment"`
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

func NewMicroserviceResources(microservice Microservice, customersTenantID string) MicroserviceResources {
	environment := strings.ToLower(microservice.Environment)
	mongoDNS := fmt.Sprintf("%s-mongo.application-%s.svc.cluster.local", environment, microservice.Application.ID)
	databasePrefix := strings.ToLower(
		fmt.Sprintf("%s_%s_%s",
			microservice.Application.Name,
			microservice.Environment,
			microservice.Name,
		))
	return MicroserviceResources{
		customersTenantID: MicroserviceResource{
			Readmodels: MicroserviceResourceReadmodels{
				Host:     fmt.Sprintf("mongodb://%s-mongo.application-%s.svc.cluster.local:27017", environment, microservice.Application.ID),
				Database: fmt.Sprintf("%s_readmodels", databasePrefix),
				Usessl:   false,
			},
			Eventstore: MicroserviceResourceStore{
				Servers: []string{
					mongoDNS,
				},
				Database: fmt.Sprintf("%s_eventstore", databasePrefix),
			},
			Projections: MicroserviceResourceStore{
				Servers: []string{
					mongoDNS,
				},
				Database: fmt.Sprintf("%s_projections", databasePrefix),
			},
			Embeddings: MicroserviceResourceStore{
				Servers: []string{
					mongoDNS,
				},
				Database: fmt.Sprintf("%s_embeddings", databasePrefix),
			},
		},
	}
}

func NewMicroserviceConfigmapPlatformData(microservice Microservice) MicroservicePlatform {
	return MicroservicePlatform{
		Applicationname:  microservice.Application.Name,
		Applicationid:    microservice.Application.ID,
		Microservicename: microservice.Name,
		Microserviceid:   microservice.ID,
		Customername:     microservice.Tenant.Name,
		Customerid:       microservice.Tenant.ID,
		Environment:      microservice.Environment,
	}
}

func NewMicroserviceConfigmap(microservice Microservice, customersTenantID string) *corev1.ConfigMap {
	name := fmt.Sprintf("%s-%s-dolittle",
		microservice.Environment,
		microservice.Name,
	)

	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)

	name = strings.ToLower(name)

	resources := NewMicroserviceResources(microservice, customersTenantID)

	endpoints := MicroserviceEndpoints{
		Public: MicroserviceEndpointPort{
			Port: 50052,
		},
		Private: MicroserviceEndpointPort{
			Port: 50053,
		},
		Management: MicroserviceEndpointPort{
			Port: 51052,
		},
	}

	metrics := MicroserviceEndpointPort{
		Port: 9700,
	}

	platform := NewMicroserviceConfigmapPlatformData(microservice)

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

	b, _ = json.MarshalIndent(metrics, "", "  ")
	metricsJSON := string(b)

	b, _ = json.MarshalIndent(platform, "", "  ")
	platformJSON := string(b)

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
			"metrics.json":                metricsJSON,
			"platform.json":               platformJSON,
		},
	}
}
