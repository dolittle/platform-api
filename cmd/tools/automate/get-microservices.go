package automate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"
)

type MicroserviceMetadataShortInfo struct {
	CustomerID       string `json:"customerId"`
	CustomerName     string `json:"customerName"`
	ApplicationID    string `json:"applicationId"`
	ApplicationName  string `json:"applicationName"`
	Environment      string `json:"environment"`
	MicroserviceID   string `json:"microserviceId"`
	MicroserviceName string `json:"microserviceName"`
}

var getMicroservicesMetaDataCMD = &cobra.Command{
	Use:   "get-microservices-metadata",
	Short: "Get all microservices metadata from the cluster",
	Long: `
	go run main.go tools get-microservices-metadata

	Outputs:
		TODO
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		kubeconfig := viper.GetString("tools.server.kubeConfig")

		if kubeconfig == "incluster" {
			kubeconfig = ""
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		microservices, err := getAllCustomerMicroservices(ctx, client)
		if err != nil {
			panic(err.Error())
		}

		data := make([]MicroserviceMetadataShortInfo, 0)
		for _, microservice := range microservices {
			data = append(data, MicroserviceMetadataShortInfo{
				ApplicationID:    microservice.Application.ID,
				ApplicationName:  microservice.Application.Name,
				Environment:      microservice.Environment,
				MicroserviceID:   microservice.ID,
				MicroserviceName: microservice.Name,
				CustomerName:     microservice.Tenant.Name,
				CustomerID:       microservice.Tenant.ID,
			})
		}

		b, _ := json.Marshal(data)
		fmt.Println(string(b))
	},
}