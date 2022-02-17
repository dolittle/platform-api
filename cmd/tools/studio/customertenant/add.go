package customertenant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var addCMD = &cobra.Command{
	Use:   "add",
	Args:  cobra.ExactArgs(1),
	Short: "Add Customer Tenant",
	Long: `


	go run main.go tools studio customer-tenant add XXX

	go run main.go tools studio customer-tenant add \
	--step=2 \
	--output=disk \
	--root-directory="/Users/freshteapot/dolittle/git/Operations" \
	--application-id="e0078604-ae62-378d-46fb-9e245d824c61" \
	--environment="Prod" \
	--microservice-id="ffb20e4f-9227-574d-31aa-d6e59b34495d" \
	22f24283-0bcf-4bbf-a8d9-da2d50bc275b


	`,
	Run: func(cmd *cobra.Command, args []string) {
		//logrus.SetFormatter(&logrus.JSONFormatter{})
		//logrus.SetOutput(os.Stdout)
		//logContext := logrus.StandardLogger()

		//applicationID := "e0078604-ae62-378d-46fb-9e245d824c61"
		//environment := "Prod"
		step, _ := cmd.Flags().GetString("step")
		customerTenantID := args[0]

		k8sClient, _ := platformK8s.InitKubernetesClient()

		// TODO write to disk
		// TODO write to cluster
		switch step {
		case "1":
			step1(cmd, k8sClient, customerTenantID)
		case "2":
			step2(cmd, k8sClient, customerTenantID)
		default:
			fmt.Println("TODO")
		}

		// var microserviceResources dolittleK8s.MicroserviceResources
		// err := json.Unmarshal([]byte(configMap.Data["resources.json"]), &microserviceResources)
	},
}

func init() {
	addCMD.Flags().String("output", "stdout", "Write to stdout, disk")
	addCMD.Flags().String("application-id", "", "Application id")
	addCMD.Flags().String("environment", "", "Environment")
	addCMD.Flags().String("microservice-id", "", "Microservice id")
	addCMD.Flags().String("step", "0", "Which step")
	addCMD.Flags().String("root-directory", "", "Root directory to write to, linked to --output=disk")
}

// Add to tenants.yaml
func step1(
	cmd *cobra.Command,
	k8sClient kubernetes.Interface,
	customerTenantID string,
) {
	outputTo, _ := cmd.Flags().GetString("output")
	rootDirectory, _ := cmd.Flags().GetString("root-directory")
	applicationID, _ := cmd.Flags().GetString("application-id")
	environment, _ := cmd.Flags().GetString("environment")

	ctx := context.TODO()
	scheme, serializer, err := automate.InitializeSchemeAndSerializer()
	if err != nil {
		panic(err.Error())
	}

	// Get tenants
	namespace := platformK8s.GetApplicationNamespace(applicationID)
	configMaps, err := automate.GetCustomerTenantsConfigMaps(ctx, k8sClient, namespace)
	if err != nil {
		panic(err)
	}

	configMap, err := automate.GetCustomerTenantsConfigMapFromConfigMaps(environment, configMaps)
	if err != nil {
		panic(err)
	}

	var runtimeTenants platform.RuntimeTenantsIDS
	json.Unmarshal([]byte(configMap.Data["tenants.json"]), &runtimeTenants)

	empty := make(map[string]interface{})
	runtimeTenants[customerTenantID] = empty

	b, _ := json.MarshalIndent(runtimeTenants, "", "  ")
	tenantsJSON := string(b)
	configMap.Data["tenants.json"] = tenantsJSON

	// Duplicated
	configMap.ManagedFields = nil
	configMap.ResourceVersion = ""
	delete(configMap.ObjectMeta.Annotations, "kubectl.kubernetes.io/last-applied-configuration")

	if outputTo == "disk" {
		err := automate.WriteCustomerTenantsToDirectory(rootDirectory, configMap)
		if err != nil {
			panic(err)
		}
		return
	}
	// Assume stdout
	automate.SetRuntimeObjectGVK(scheme, &configMap)
	serializer.Encode(&configMap, os.Stdout)
}

// Add customer tenant to resource.json
func step2(
	cmd *cobra.Command,
	k8sClient kubernetes.Interface,
	customerTenantID string,
) {
	//outputTo, _ := cmd.Flags().GetString("output")
	//rootDirectory, _ := cmd.Flags().GetString("root-directory")
	applicationID, _ := cmd.Flags().GetString("application-id")
	environment, _ := cmd.Flags().GetString("environment")
	microserviceID, _ := cmd.Flags().GetString("microservice-id")

	ctx := context.TODO()

	// Get tenants

	configMap, err := automate.GetDolittleConfigMap(ctx, k8sClient, applicationID, environment, microserviceID)

	if err != nil {
		panic(err)
	}

	var microserviceResources dolittleK8s.MicroserviceResources
	err = json.Unmarshal([]byte(configMap.Data["resources.json"]), &microserviceResources)
	if err != nil {
		panic(err)
	}

	newResources := dolittleK8s.NewMicroserviceResources(applicationID, environment, microserviceID, []platform.CustomerTenantInfo{
		{
			CustomerTenantID: customerTenantID,
		},
	})
	newResource := newResources[customerTenantID]

	b, _ := json.MarshalIndent(newResource, "", "  ")
	fmt.Println(string(b))
	//microserviceResources[customerTenantID] = newResource
	//
	//// TODO might need to do some extra work here to avoid existing data = null
	//// Do we try and fix it?
	//b, _ := json.MarshalIndent(microserviceResources, "", "  ")
	//configMap.Data["resources.json"] = string(b)

	// Duplicated
	//scheme, serializer, err := automate.InitializeSchemeAndSerializer()
	//if err != nil {
	//	panic(err.Error())
	//}

	//configMap.ManagedFields = nil
	//configMap.ResourceVersion = ""
	//delete(configMap.ObjectMeta.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
	//
	//// TODO figure out how to write to disk
	//// not ready yet
	//if outputTo == "disk" {
	//	err := automate.WriteConfigMapsToDirectory(rootDirectory, []corev1.ConfigMap{*configMap})
	//	if err != nil {
	//		panic(err)
	//	}
	//	return
	//}
	//// Assume stdout
	//automate.SetRuntimeObjectGVK(scheme, configMap)
	//serializer.Encode(configMap, os.Stdout)

}
