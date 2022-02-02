package studio

import (
	"context"
	"fmt"
	"os"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO should we deprecate this? or make it reusable in terms of add the "name" and it will hook up the developer rbac
var createServiceAccountCMD = &cobra.Command{
	Use:   "create-service-account",
	Short: "Create a k8s devops service account for an application",
	Long: `
	Attempts to create a "devops" serviceaccount for the application and adds it to the already existing "developer" rolebinding.

	go run main.go tools studio create-service-account --all
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()

		ctx := context.TODO()

		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()
		k8sRepo := platformK8s.NewK8sRepo(k8sClient, k8sConfig, logContext.WithField("context", "k8s-repo"))

		createAll, _ := cmd.Flags().GetBool("all")
		if createAll && len(args) > 0 {
			logContext.Fatal("Specify either the APPLICATIONID or the '--all' flag")
		}

		addedAccounts := 0

		serviceAccount := "devops"
		roleBinding := "devops"

		if createAll {
			logContext.Info("Adding a devops service account for all applications")
			applications := extractApplications(ctx, k8sClient)

			for _, application := range applications {
				err := k8sRepo.AddServiceAccount(serviceAccount, roleBinding, application.TenantID, application.TenantName, application.ID, application.Name)
				if err != nil {
					if err != platformK8s.ErrAlreadyExists {
						panic(err.Error())
					}
					logContext.Infof("Application '%s' already had the service account or rolebinding, skipping", application.ID)
					// the account already existed or it already had a rolebinding so don't increment
					continue
				}
				logContext.Infof("Added a service account for application %s", application.ID)
				addedAccounts++
			}
			logContext.Infof("Added %v service accounts", addedAccounts)
		} else {
			if len(args) < 1 {
				logContext.Fatal("Specify the APPLICATIONID or the '--all' flag")
			}

			applicationID := args[0]

			namespace := fmt.Sprintf("application-%s", applicationID)
			k8sNamespace, err := k8sClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
			if err != nil {
				logContext.Fatalf("Couldn't find the specified namespace: %s", namespace)
			}

			customerID := k8sNamespace.Annotations["dolittle.io/tenant-id"]
			customerName := k8sNamespace.Labels["tenant"]
			applicationName := k8sNamespace.Labels["application"]

			err = k8sRepo.AddServiceAccount(serviceAccount, roleBinding, customerID, customerName, applicationID, applicationName)

			if err != nil {
				if err != platformK8s.ErrAlreadyExists {
					panic(err.Error())
				}
				logContext.Infof("Application '%s' already had the service account or rolebinding, skipping", applicationID)
			}
		}
		logContext.Info("Finished!")
	},
}

func init() {
	createServiceAccountCMD.Flags().Bool("all", false, "Add a devops serviceaccount for all customers")
}
