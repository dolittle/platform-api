package microservice

import (
	"context"
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var createServiceAccountCMD = &cobra.Command{
	Use:   "create-service-account",
	Short: "Create a k8s devops service account for an application",
	Long: `
	Attempts to create a "devops" serviceaccount for the application and adds it to the already existing "developer" rolebinding.

	go run main.go microservice create-service-account --all
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()

		ctx := context.TODO()
		kubeconfig := viper.GetString("tools.server.kubeConfig")

		if kubeconfig == "incluster" {
			kubeconfig = ""
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		k8sRepo := platform.NewK8sRepo(client, config)

		createAll := viper.GetBool("all")
		if createAll && len(args) > 0 {
			logContext.Fatal("Specify either the APPLICATIONID or the '--all' flag")
		}

		addedAccounts := 0

		if createAll {
			logContext.Info("Adding a devops service account for all applications")
			applications := extractApplications(ctx, client)

			for _, application := range applications {
				err := addServiceAccount(logContext, k8sRepo, application.TenantID, application.TenantName, application.ID, application.Name)
				if err != nil {
					if err != platform.AlreadyExists {
						panic(err.Error())
					}
					logContext.Infof("Application '%s' already had the service account or rolebinding, skipping", application.ID)
					// the account already existed or it already had a rolebinding so don't increment
					continue
				}
				addedAccounts++
			}
			logContext.Infof("Added %v service accounts", addedAccounts)
		} else {
			if len(args) < 1 {
				logContext.Fatal("Specify the APPLICATIONID or the '--all' flag")
			}

			applicationID := args[0]

			namespace := fmt.Sprintf("application-%s", applicationID)
			k8sNamespace, err := client.CoreV1().Namespaces().Get(ctx, namespace, v1.GetOptions{})
			if err != nil {
				logContext.Fatalf("Couldn't find the specified namespace: %s", namespace)
			}

			customerID := k8sNamespace.Annotations["dolittle.io/tenant-id"]
			customerName := k8sNamespace.Labels["tenant"]
			applicationName := k8sNamespace.Labels["application"]
			err = addServiceAccount(logContext, k8sRepo, customerID, customerName, applicationID, applicationName)
			if err != nil {
				if err != platform.AlreadyExists {
					panic(err.Error())
				}
				logContext.Infof("Application '%s' already had the service account or rolebinding, skipping", applicationID)
			}
		}
		logContext.Info("Finished!")
	},
}

func addServiceAccount(logger logrus.FieldLogger, k8sRepo platform.K8sRepo, customerID string, customerName string, applicationID string, applicationName string) error {
	serviceAccount := "devops"
	roleBinding := "devops"
	logContext := logger.WithFields(logrus.Fields{
		"customerID":     customerID,
		"applicationID":  applicationID,
		"serviceAccount": serviceAccount,
		"roleBinding":    roleBinding,
		"function":       "createServiceAccount",
	})

	_, err := k8sRepo.CreateServiceAccount(logContext,
		customerID,
		customerName,
		applicationID,
		applicationName,
		serviceAccount,
	)

	if err != nil {
		if err != platform.AlreadyExists {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("failed to create the devops serviceaccount")
		}
		return err
	}

	_, err = k8sRepo.CreateRoleBinding(
		logContext,
		customerID,
		customerName,
		applicationID,
		applicationName,
		roleBinding,
		"developer",
	)
	if err != nil && err != platform.AlreadyExists {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to create the rolebinding")
		return err
	}

	_, err = k8sRepo.AddServiceAccountToRoleBinding(logContext, applicationID, roleBinding, serviceAccount)
	if err != nil {
		if err != platform.AlreadyExists {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("failed to add the service account to the rolebinding")
		}
		return err
	}

	logContext.Infof("Added a service account for application %s", applicationID)

	return nil
}

func init() {
	RootCmd.AddCommand(createServiceAccountCMD)

	createServiceAccountCMD.Flags().Bool("all", false, "Add a devops serviceaccount for all customers")
	viper.BindPFlag("all", createServiceAccountCMD.Flags().Lookup("all"))
}
