package create

import (
	"encoding/json"
	"fmt"
	"os"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/welcome"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	networkingv1 "k8s.io/api/networking/v1"
)

// TODO
// - file is not in use yet
var microserviceCMD = &cobra.Command{
	Use:   "microservice",
	Short: "Create a microservice",
	Long: `
	Create a micorservice via the command line.

	Stub a template:

	go run main.go tools studio create microservice --template
	go run main.go tools studio create microservice --template --customer-id="66587d28-d14c-424a-b3dd-f24a22f586bf" --application-id="not-real" --environment="Dev" --microservice-id="fake-ms"


	go run main.go tools studio create microservice --template 66587d28-d14c-424a-b3dd-f24a22f586bf/not-real/Dev/new
	go run main.go tools studio create microservice --template 66587d28-d14c-424a-b3dd-f24a22f586bf/not-real/Dev/fake-ms


	go run main.go tools studio create microservice -f microservice.json
	`,
	Run: func(cmd *cobra.Command, args []string) {

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

		customerID, _ := cmd.Flags().GetString("customer-id")
		applicationID, _ := cmd.Flags().GetString("application-id")
		environment, _ := cmd.Flags().GetString("environment")
		microserviceID, _ := cmd.Flags().GetString("microservice-id")

		template, _ := cmd.Flags().GetBool("template")
		if template {
			printMicoserviceJSON(customerID, applicationID, environment, microserviceID)
			return
		}

		pathToFile, _ := cmd.Flags().GetString("f")
		if pathToFile == "" {
			fmt.Println("-f is empty")
			return
		}

		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		isProduction := viper.GetBool("tools.server.isProduction")

		gitRepoConfig := git.InitGit(logger, platformEnvironment)
		// TODO until we fix the git pull issue, I am not sure this will work without a restart.
		storageRepo := gitStorage.NewGitStorage(
			logger.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		// What runtime image?
		// What name?
		// What environment?
		// Or import and override?

		headImage := "welcome"
		// This might need to be the same in the application?
		microserviceName := ""
		runtimeImage := "TODO"
		urlPath := "/todo"
		namesapce := "TODO"

		// Simpler with one for now

		environments := []string{
			environment,
		}

		logContext := logger.WithFields(logrus.Fields{
			"customer_id":     customerID,
			"application_id":  applicationID,
			"environments":    environments,
			"microservice_id": microserviceID,
		})

		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()
		k8sRepo := platformK8s.NewK8sRepo(k8sClient, k8sConfig, logContext.WithField("context", "k8s-repo"))
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		microserviceSimpleRepo := k8sSimple.NewSimpleRepo(k8sClient, k8sRepo, k8sRepoV2, isProduction)

		// Build microservice
		// Not sold on this approach, but its the code that exists now
		// requestBytes, microserviceBase, err := s.readMicroserviceBase(request, logContext)
		// Confirm application exists
		studioInfo, err := storage.GetStudioInfo(storageRepo, customerID, applicationID, logContext)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		tenantInfo := dolittleK8s.Tenant{
			Name: studioInfo.TerraformCustomer.Name,
			ID:   studioInfo.TerraformCustomer.GUID,
		}

		applicationInfo := dolittleK8s.Application{
			Name: studioInfo.TerraformApplication.Name,
			ID:   studioInfo.TerraformApplication.ApplicationID,
		}

		application, err := storageRepo.GetApplication(customerID, applicationInfo.ID)
		if err != nil {
			panic(err.Error())
		}

		newMicroservice := platform.HttpInputSimpleInfo{
			MicroserviceBase: platform.MicroserviceBase{
				Dolittle: platform.HttpInputDolittle{
					ApplicationID:  applicationID,
					CustomerID:     customerID,
					MicroserviceID: microserviceID,
				},
				Name: microserviceName,
				Kind: platform.MicroserviceKindSimple,
			},
			Extra: platform.HttpInputSimpleExtra{
				Headimage:    headImage,
				Runtimeimage: runtimeImage,
				Ingress: platform.HttpInputSimpleIngress{
					Path:     urlPath,
					Pathtype: string(networkingv1.PathTypePrefix),
				},
			},
		}

		// Loop environments
		created := make([]string, 0)
		for _, environment := range application.Environments {
			if !funk.Contains(environments, environment.Name) {
				continue
			}

			newMicroservice.Environment = environment.Name
			//if dryRun {
			//  resources := k8sSimple.NewResources(isProduction, namesapce, tenantInfo, applicationInfo, environment.CustomerTenants, newMicroservice)
			//	// How to write this to disk?
			//	// I can't fully as somethings change the rbac
			//	fmt.Println("Write to disk?", resources.ConfigEnvironmentVariables)
			//	fmt.Println("Write to disk?", resources.SecretEnvironmentVariables)
			//	fmt.Println("Write to disk?", resources.ConfigFiles)
			//	fmt.Println("Write to disk?", resources.Deployment)
			//	fmt.Println("Write to disk?", resources.Ingresses)
			//	fmt.Println("Write to disk?", resources.Service)
			//	fmt.Println("Write to disk?", resources.DolittleConfig)
			//	fmt.Println("Write to disk?", resources.NetworkPolicy)
			//	fmt.Println("Write to disk?", resources.RbacPolicyRules)
			//	continue
			//}

			// TODO write to disk
			err = storageRepo.SaveMicroservice(
				newMicroservice.Dolittle.CustomerID,
				newMicroservice.Dolittle.ApplicationID,
				newMicroservice.Environment,
				newMicroservice.Dolittle.MicroserviceID,
				newMicroservice,
			)

			if err != nil {
				fmt.Println("Failed to save to disk")
				fmt.Println(err.Error())
				return
			}

			// TODO build
			err = microserviceSimpleRepo.Create(namesapce, tenantInfo, applicationInfo, environment.CustomerTenants, newMicroservice)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			created = append(created, environment.Name)
		}

		// TODO what is the command to write to disk?
		logContext.WithFields(logrus.Fields{
			"created": created,
		}).Info("Finished!")
	},
}

func init() {
	microserviceCMD.Flags().Bool("template", false, "Print microservice json template")
	microserviceCMD.Flags().String("f", "", "path to microservice json file")
	microserviceCMD.Flags().String("customer-id", "", "customer id")
	microserviceCMD.Flags().String("application-id", "", "application id")
	microserviceCMD.Flags().String("environment", "", "environment")
	microserviceCMD.Flags().String("microservice-id", "", "microservice id")
}

func printMicoserviceJSON(customerID string, applicationID string, environment string, microserviceID string) {
	headImage := welcome.Image
	runtimeImage := "dolittle/runtime:7.8.0"
	sample := platform.HttpInputSimpleInfo{
		MicroserviceBase: platform.MicroserviceBase{
			Dolittle: platform.HttpInputDolittle{
				ApplicationID:  applicationID,
				CustomerID:     customerID,
				MicroserviceID: microserviceID,
			},
			Name:        "microserviceName",
			Environment: environment,
			Kind:        platform.MicroserviceKindSimple,
		},
		Extra: platform.HttpInputSimpleExtra{
			Headimage:    headImage,
			Runtimeimage: runtimeImage,
			Ingress: platform.HttpInputSimpleIngress{
				Path:     "/change",
				Pathtype: string(networkingv1.PathTypePrefix),
			},
		},
	}
	b, _ := json.MarshalIndent(sample, "", "  ")
	fmt.Println(string(b))
}
