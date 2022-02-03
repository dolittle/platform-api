package explore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
)

var ingressCMD = &cobra.Command{
	Use:   "ingress",
	Short: "Get ingress data for customer tenants + microservices",
	Long: `
	go run main.go tools explore ingress
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()
		logger.Info("Hello")

		ctx := context.TODO()
		k8sClient, _ := platformK8s.InitKubernetesClient()

		namespaces := automate.GetNamespacesWithApplication(ctx, k8sClient)

		for _, namespace := range namespaces {

			ingresses, err := automate.GetIngresses(ctx, k8sClient, namespace.Name)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Getting ingresses")
			}
			for _, ingress := range ingresses {
				tenantHeaderAnnotation := ingress.GetObjectMeta().GetAnnotations()["nginx.ingress.kubernetes.io/configuration-snippet"]
				customerTenantID := platformK8s.GetCustomerTenantIDFromNginxConfigurationSnippet(tenantHeaderAnnotation)
				if customerTenantID == "" {
					continue
				}

				data := map[string]string{
					"namesspace":         namespace.Name,
					"customer_id":        ingress.Annotations["dolittle.io/tenant-id"],
					"application_id":     ingress.Annotations["dolittle.io/application-id"],
					"microservice_id":    ingress.Annotations["dolittle.io/microservice-id"],
					"customer_tenant_id": customerTenantID,
					"environent":         ingress.Labels["environment"],
					"ingress_namee":      ingress.Name,
					"cmd": fmt.Sprintf(
						`kubectl -n %s get ingress %s -oyaml`,
						namespace.Name,
						ingress.Name,
					),
				}

				for _, rule := range ingress.Spec.Rules {

					for _, ingressPath := range rule.HTTP.Paths {
						data["path"] = ingressPath.Path
						data["host"] = rule.Host

						b, _ := json.Marshal(data)
						fmt.Println(string(b))
					}
				}

			}

		}
	},
}
