package automate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	"k8s.io/client-go/tools/clientcmd"
)

var getMicroservicesMetaDataCMD = &cobra.Command{
	Use:   "get-microservices-metadata",
	Short: "Get all microservices metadata from the cluster",
	Long: `
	go run main.go tools automate get-microservices-metadata

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

		microservices, err := automate.GetAllCustomerMicroservices(ctx, client)
		if err != nil {
			panic(err.Error())
		}

		data := make([]platform.MicroserviceMetadataShortInfo, 0)
		for _, microservice := range microservices {
			data = append(data, platform.MicroserviceMetadataShortInfo{
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
