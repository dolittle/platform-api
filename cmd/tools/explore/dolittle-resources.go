package explore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
)

var dolittleResourcesCMD = &cobra.Command{
	Use:   "dolittle-resources",
	Short: "Get dolittle resources",
	Long: `
	go run main.go tools explore dolittle-resources


	Explore data that is broken	
	go run main.go tools explore dolittle-resources | grep '{' | jq -r '.tips'

	Explore data that exists
	go run main.go tools explore dolittle-resources | grep -v '^{'
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()
		logContext := logger.WithField("cmd", "dolittle-resources")

		k8sClient, _ := platformK8s.InitKubernetesClient()
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		namespaces, err := k8sRepoV2.GetNamespacesWithApplication()

		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Getting namespaces")
		}

		ctx := context.TODO()
		for _, namespace := range namespaces {
			logContext = logContext.WithFields(logrus.Fields{
				"namespace": namespace.Name,
			})

			configMaps, err := automate.GetDolittleConfigMaps(ctx, k8sClient, namespace.Name)
			if err != nil {
				logContext.Fatal("Failed to get *-dolittle configmaps")
			}

			for _, configMap := range configMaps {
				logContext = logContext.WithFields(logrus.Fields{
					"name": configMap.Name,
				})

				var microserviceResources dolittleK8s.MicroserviceResources
				err := json.Unmarshal([]byte(configMap.Data["resources.json"]), &microserviceResources)
				if err != nil {
					logContext.WithFields(logrus.Fields{
						"error": err,
						"tips": fmt.Sprintf(
							`kubectl -n %s get configmaps %s -ojson | jq '.data'`,
							configMap.Namespace,
							configMap.Name,
						),
					}).Error("Failed to parse resources.json")
					continue
				}

				for customerTenantID, microserviceResource := range microserviceResources {
					fmt.Println(customerTenantID, microserviceResource.Readmodels.Database)
				}
			}

		}
	},
}
