package explore

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/dolittle/platform-api/pkg/k8s"
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

		logContext := logrus.StandardLogger()
		logContext.Info("Hello")

		k8sClient, _ := platformK8s.InitKubernetesClient()

		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		namespaces, err := k8sRepoV2.GetNamespacesWithApplication()
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Getting namespaces")
		}

		for _, namespace := range namespaces {

			ingresses, err := k8sRepoV2.GetIngresses(namespace.Name)
			if err != nil {
				logContext.WithFields(logrus.Fields{
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
