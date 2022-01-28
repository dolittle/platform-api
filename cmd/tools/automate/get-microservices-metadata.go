package automate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
)

var getMicroservicesMetaDataCMD = &cobra.Command{
	Use:   "get-microservices-metadata",
	Short: "Get all microservices metadata from the cluster",
	Long: `
go run main.go tools automate get-microservices-metadata

Returns an array of metadata

Outputs:
	Example of the metadata
	{
		"customerId": "508c1745-5f2a-4b4c-b7a5-2fbb1484346d",
		"customerName": "Dolittle",
		"applicationId": "fe7736bb-57fc-4166-bb91-6954f4dd4eb7",
		"applicationName": "Studio",
		"environment": "Dev",
		"microserviceId": "f966abb5-3d22-4c2d-b5cb-1c49e5946e03",
		"microserviceName": "SelfServiceWeb"
	}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, _ := platformK8s.InitKubernetesClient()

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
