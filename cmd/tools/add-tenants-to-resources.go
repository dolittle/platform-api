package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	// "k8s.io/apimachinery/pkg/runtime"
	// k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

type dolittleConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Annotations struct {
			DolittleIoTenantID       string `yaml:"dolittle.io/tenant-id"`
			DolittleIoApplicationID  string `yaml:"dolittle.io/application-id"`
			DolittleIoMicroserviceID string `yaml:"dolittle.io/microservice-id"`
		} `yaml:"annotations"`
		Labels struct {
			Tenant       string `yaml:"tenant"`
			Application  string `yaml:"application"`
			Environment  string `yaml:"environment"`
			Microservice string `yaml:"microservice"`
		} `yaml:"labels"`
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Data struct {
		ResourcesJSON            string `yaml:"resources.json"`
		EventHorizonsJSON        string `yaml:"event-horizons.json"`
		EventHorizonConsentsJSON string `yaml:"event-horizon-consents.json"`
		MicroservicesJSON        string `yaml:"microservices.json"`
		EndpointsJSON            string `yaml:"endpoints.json"`
		AppsettingsJSON          string `yaml:"appsettings.json"`
	} `yaml:"data"`
}

type resourceJSON map[string]tenantResource

type tenantResource struct {
	ReadModels ReadModels `json:"readModels"`
	EventStore EventStore `json:"eventStore"`
}
type ReadModels struct {
	Host     string `json:"host"`
	Database string `json:"database"`
	UseSSL   bool   `json:"useSSL"`
}
type EventStore struct {
	Servers               []string `json:"servers"`
	Database              string   `json:"database"`
	MaxConnectionPoolSize int      `json:"maxConnectionPoolSize"`
}

var addTenantsToResourcesCMD = &cobra.Command{
	Use:   "add-tenants-to-resources",
	Short: "Adds a new tenant to a microservices k8s resource file",
	Long: `
	
	`,
	Run: func(cmd *cobra.Command, args []string) {

		// runtimeScheme := runtime.NewScheme()
		// serializer := k8sJson.NewSerializerWithOptions(
		// 	k8sJson.DefaultMetaFactory,
		// 	runtimeScheme,
		// 	runtimeScheme,
		// 	k8sJson.SerializerOptions{
		// 		Yaml:   true,
		// 		Pretty: true,
		// 		Strict: true,
		// 	},
		// )

		if len(args) == 0 {
			fmt.Println("You have to provide a filename")
			return
		}
		fileName := args[0]

		data, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println("Couldn't read file")
			panic(err)
		}

		config := &dolittleConfig{}
		err = yaml.Unmarshal(data, config)
		if err != nil {
			fmt.Println("Couldn't unmarshal yaml file")
			panic(err)
		}

		microserviceID := config.Metadata.Annotations.DolittleIoMicroserviceID[0:8]

		resourcesJson := make(resourceJSON)

		err = json.Unmarshal([]byte(config.Data.ResourcesJSON), &resourcesJson)
		if err != nil {
			panic(err)
		}

		fmt.Printf("There's %v tenant(s) \n", len(resourcesJson))

		var dbHost string
		// get the first host from the resources.json
		for _, value := range resourcesJson {
			dbHost = value.EventStore.Servers[0]
			break
		}

		for i := 0; i < 9; i++ {

			newTenant := uuid.New().String()

			newDatabase := fmt.Sprintf("%s_%s", microserviceID, newTenant[0:8])

			readmodels := fmt.Sprintf("%s_readmodels", newDatabase)
			eventstore := fmt.Sprintf("%s_eventstore", newDatabase)

			resourcesJson[newTenant] = tenantResource{
				ReadModels: ReadModels{
					Host:     fmt.Sprintf("mongodb://%s:27017", dbHost),
					Database: readmodels,
					UseSSL:   false,
				},
				EventStore: EventStore{
					Servers: []string{
						dbHost,
					},
					Database:              eventstore,
					MaxConnectionPoolSize: 2000,
				},
			}

			fmt.Printf("TenantID: %s Database: %s Eventstore: %s\n", newTenant, readmodels, eventstore)
		}
		fmt.Printf("Added new tenant(s), new amount of tenants is %v\n", len(resourcesJson))

		resourcesJsonString, err := json.MarshalIndent(resourcesJson, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println("Here's the new resources.json, copy and paste it:")
		fmt.Println(string(resourcesJsonString))
		// config.Data.ResourcesJSON = string(resourcesJsonString)

		// newYaml, err := yaml.Marshal(config)
		// if err != nil {
		// 	panic(err)
		// }

		// fmt.Println("printing the new yaml")
		// fmt.Println(string(newYaml))

		// serializer.Encode(config, os.Stdout)
	},
}
