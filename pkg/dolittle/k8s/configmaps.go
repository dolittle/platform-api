package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type emptyObject struct{}

type MicroserviceResources map[string]MicroserviceResource

// MicroserviceResource
/*
Today we have custom variables for the respective  databases
Going forward we are looking to use a key based on the microserviceID and the customerTenantID
At the moment we do not need to know the individual information as we can consume the data as
one whole object.
The main reason for the new approach is to create a repeatable approach with a very low level of collision for the prefix
*/
type MicroserviceResource struct {
	Readmodels  MicroserviceResourceReadmodels `json:"readModels"`
	Eventstore  MicroserviceResourceStore      `json:"eventStore"`
	Projections MicroserviceResourceStore      `json:"projections"`
	Embeddings  MicroserviceResourceStore      `json:"embeddings"`
}
type MicroserviceResourceReadmodels struct {
	Host     string `json:"host"`
	Database string `json:"database"`
	UseSSL   bool   `json:"useSSL"`
}
type MicroserviceResourceStore struct {
	Servers  []string `json:"servers"`
	Database string   `json:"database"`
}

type MicroserviceEndpointsV6_1_0 struct {
	Public  MicroserviceEndpointPort `json:"public"`
	Private MicroserviceEndpointPort `json:"private"`
}

type MicroserviceEndpoints struct {
	Public     MicroserviceEndpointPort `json:"public"`
	Private    MicroserviceEndpointPort `json:"private"`
	Management MicroserviceEndpointPort `json:"management"`
}

type MicroserviceEndpointPort struct {
	Port int `json:"port"`
}

// event-horizon-consents.json
type MicroserviceEventHorizonConsents map[string][]MicroserviceConsent

type MicroserviceConsent struct {
	Microservice string `json:"microservice"`
	Tenant       string `json:"tenant"`
	Stream       string `json:"stream"`
	Partition    string `json:"partition"`
	Consent      string `json:"consent"`
}

// microservices.json
type MicroserviceMicroservices map[string]MicroserviceMicroservice

type MicroserviceMicroservice struct {
	Host string `json:"host"`
	Port int32  `json:"port"`
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

// AppsettingsV8_0_0 follows the structure of DOLITTLE__RUNTIME__EVENTSTORE__BACKWARDSCOMPATIBILITY__VERSION env variable
// so that the ASP.NET configuration system will pick this setting in appsettings.json as that env variable
type AppsettingsV8_0_0 struct {
	Appsettings
	Dolittle dolittle `json:"dolittle"`
}
type dolittle struct {
	Runtime runtime `json:"runtime"`
}
type runtime struct {
	EventStore eventStore `json:"eventstore"`
}
type eventStore struct {
	BackwardsCompatibility backwardsCompatibility `json:"backwardscompatibility"`
}
type backwardsCompatibility struct {
	Version BackwardsCompatibilityVersion `json:"version"`
}

type BackwardsCompatibilityVersion string

const (
	V6BackwardsCompatibility BackwardsCompatibilityVersion = "V6"
	V7BackwardsCompatibility BackwardsCompatibilityVersion = "V7"
)

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

// ResourcePrefix Create a uniq preifx
// Linked to resources.json inside *-dolittle configmap
func ResourcePrefix(microserviceID string, customerTenantID string) string {
	return strings.ToLower(
		fmt.Sprintf("%s_%s",
			microserviceID[0:7],
			customerTenantID[0:7],
		))
}

// NewMicroserviceResources
// Build the microservice resource creating custmoer tenants specific blocks
func NewMicroserviceResources(microservice Microservice, customerTenants []platform.CustomerTenantInfo) MicroserviceResources {

	environment := strings.ToLower(microservice.Environment)
	mongoDNS := fmt.Sprintf("%s-mongo.application-%s.svc.cluster.local", environment, microservice.Application.ID)

	resources := MicroserviceResources{}

	for _, customerTenant := range customerTenants {
		customerTenantID := customerTenant.CustomerTenantID
		databasePrefix := ResourcePrefix(microservice.ID, customerTenant.CustomerTenantID)

		dolittleResource := MicroserviceResource{
			Readmodels: MicroserviceResourceReadmodels{
				Host:     fmt.Sprintf("mongodb://%s-mongo.application-%s.svc.cluster.local:27017", environment, microservice.Application.ID),
				Database: fmt.Sprintf("%s_readmodels", databasePrefix),
				UseSSL:   false,
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
		}
		resources[customerTenantID] = dolittleResource
	}

	return resources
}

func NewMicroserviceConfigMapPlatformData(microservice Microservice) MicroservicePlatform {
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

// NewMicroserviceConfigmap create dolittle-config configmap specific for dolittle/runtime:6.1.0
func NewMicroserviceConfigmapV6_1_0(microservice Microservice, customersTenants []platform.CustomerTenantInfo) *corev1.ConfigMap {
	configmap := NewMicroserviceConfigmap(microservice, customersTenants)

	endpoints := MicroserviceEndpointsV6_1_0{
		Public: MicroserviceEndpointPort{
			Port: 50052,
		},
		Private: MicroserviceEndpointPort{
			Port: 50053,
		},
	}

	b, _ := json.MarshalIndent(endpoints, "", "  ")
	endpointsJSON := string(b)

	configmap.Data["endpoints.json"] = endpointsJSON
	return configmap
}

// NewMicroserviceConfigmap create dolittle-config configmap
func NewMicroserviceConfigmap(microservice Microservice, customersTenants []platform.CustomerTenantInfo) *corev1.ConfigMap {
	name := fmt.Sprintf("%s-%s-dolittle",
		microservice.Environment,
		microservice.Name,
	)

	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)

	name = strings.ToLower(name)

	resources := NewMicroserviceResources(microservice, customersTenants)

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

	platform := NewMicroserviceConfigMapPlatformData(microservice)

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

	eventHorizonConsents := MicroserviceEventHorizonConsents{}
	b, _ = json.MarshalIndent(eventHorizonConsents, "", "  ")
	eventHorizonConsentsJSON := string(b)

	microservices := MicroserviceMicroservices{}
	b, _ = json.MarshalIndent(microservices, "", "  ")
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

// NewMicroserviceConfigmap creates dolittle-config configmap specific for dolittle/runtime:8.0.0
// where the backwardsCompatibility is set to V7. The backwardsCompatibility is set in the Runtime
// with the DOLITTLE__RUNTIME__EVENTSTORE__BACKWARDSCOMPATIBILITY__VERSION env variable, but instead
// of using an environment variable we can leverage ASP.NET appsettings.json file to set it.
// We set it to V7 by default as the V6 option should only be used when upgrading the Runtime, while this deals
// with only creating a new one. https://github.com/dolittle/Runtime/releases/tag/v8.0.0
func NewMicroserviceConfigmapV8_0_0(microservice Microservice, customersTenants []platform.CustomerTenantInfo) *corev1.ConfigMap {
	configmap := NewMicroserviceConfigmap(microservice, customersTenants)

	appsettings := AppsettingsV8_0_0{
		Appsettings: Appsettings{
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
		},
		Dolittle: dolittle{
			Runtime: runtime{
				EventStore: eventStore{
					BackwardsCompatibility: backwardsCompatibility{
						Version: V7BackwardsCompatibility,
					},
				},
			},
		},
	}

	b, _ := json.MarshalIndent(appsettings, "", "  ")
	appsettingsJSON := string(b)

	configmap.Data["appsettings.json"] = appsettingsJSON
	return configmap
}
