package microservice

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
)

// microserviceCmd represents the microservice command
var createCMD = &cobra.Command{
	Use:   "create",
	Short: "Create a microservice",
	Long: `

	go run main.go microservice > data.ndjson
	kubectl diff -f data.ndjson

TODO:
	- configmaps
	- secrets
	- service
	- UI
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO how do you get a name for ingress?
		// Possible input
		microserviceID := "9f6a613f-d969-4938-a1ac-5b7df199bc40"
		customersTenantID := "17426336-fb8e-4425-8ab7-07d488367be9"
		customerTenants := []platform.CustomerTenantInfo{
			{
				CustomerTenantID: customersTenantID,
				Hosts: []platform.CustomerTenantHost{
					{
						Host:       "freshteapot-taco.dolittle.cloud",
						SecretName: "freshteapot-taco-certificate",
					},
				},
			},
		}
		headImage := "453e04a74f9d42f2b36cd51fa2c83fa3.azurecr.io/taco/order:1.0.6"
		runtimeImage := "dolittle/runtime:5.3.3"
		microservice := k8s.Microservice{
			ID:   microserviceID,
			Name: "Order2",
			Tenant: k8s.Tenant{
				ID:   "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3",
				Name: "Customer-Chris",
			},
			Application: k8s.Application{
				ID:   "11b6cf47-5d9f-438f-8116-0d9828654657",
				Name: "Taco",
			},
			Environment: "Dev",
		}

		host := "freshteapot-taco.dolittle.cloud"
		secretName := "freshteapot-taco-certificate"

		ingressServiceName := strings.ToLower(fmt.Sprintf("%s-%s", microservice.Environment, microservice.Name))
		ingressRules := []k8s.SimpleIngressRule{
			{
				Path:            "/test",
				PathType:        networkingv1.PathType("Prefix"),
				ServiceName:     ingressServiceName,
				ServicePortName: "http",
			},
		}

		platformEnvironment := "dev"

		// Dolittle micoservice
		microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, customerTenants)
		deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
		service := k8s.NewService(microservice)
		ingress := k8s.NewMicroserviceIngressWithEmptyRules(platformEnvironment, microservice)
		configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
		configFiles := k8s.NewConfigFilesConfigmap(microservice)
		configSecrets := k8s.NewEnvVariablesSecret(microservice)

		ingress = k8s.AddCustomerTenantIDToIngress(customersTenantID, ingress)
		ingress.Spec.TLS = k8s.AddIngressTLS([]string{host}, secretName)
		ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.AddIngressRule(host, ingressRules))

		// One per line is the magic
		data := make([]interface{}, 0)
		data = append(data, deployment)
		data = append(data, service)
		data = append(data, ingress)
		data = append(data, microserviceConfigmap)
		data = append(data, configEnvVariables)
		data = append(data, configFiles)
		data = append(data, configSecrets)

		for _, resource := range data {
			b, _ := json.Marshal(resource)
			//yaml.JSONToYAML()
			//b, _ := yaml.Marshal(resource)
			fmt.Println(string(b))
		}
	},
}
