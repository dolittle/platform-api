package customertenant

import (
	"context"
	"encoding/json"
	"os"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var addCMD = &cobra.Command{
	Use:   "add",
	Args:  cobra.ExactArgs(2),
	Short: "Add Customer Tenant",
	Long: `


	go run main.go tools studio customer-tenant add XXX
	`,
	Run: func(cmd *cobra.Command, args []string) {
		//logrus.SetFormatter(&logrus.JSONFormatter{})
		//logrus.SetOutput(os.Stdout)
		//logContext := logrus.StandardLogger()

		outputTo, _ := cmd.Flags().GetString("output")

		applicationID := "e0078604-ae62-378d-46fb-9e245d824c61"
		environment := "Prod"
		customerTenantID := args[0]
		rootDirectory := args[1]
		///Users/freshteapot/dolittle/git/Operations
		//fmt.Println(logContext, applicationID, environment, customerTenantID)

		k8sClient, _ := platformK8s.InitKubernetesClient()

		// TODO write to disk
		// TODO write to cluster
		step1(k8sClient, applicationID, environment, customerTenantID, rootDirectory, outputTo)

		// var microserviceResources dolittleK8s.MicroserviceResources
		// err := json.Unmarshal([]byte(configMap.Data["resources.json"]), &microserviceResources)
	},
}

func init() {
	addCMD.PersistentFlags().String("output", "stdout", "Write to stdout, disk")
}

// Add to tenants.yaml
func step1(
	k8sClient kubernetes.Interface,
	applicationID string,
	environment string,
	customerTenantID string,
	rootDirectory string,
	outputTo string,
) {
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
