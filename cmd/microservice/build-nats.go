package microservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

var buildNatsCMD = &cobra.Command{
	Use:   "build-nats",
	Short: "Building nats",
	Long: `
	go run main.go microservice build-nats --kube-config="/Users/freshteapot/.kube/config" /tmp/single-server-nats.yml
	`,
	Run: func(cmd *cobra.Command, args []string) {
		gitRepoBranch := viper.GetString("tools.server.gitRepo.branch")
		if gitRepoBranch == "" {
			panic("GIT_BRANCH required")
		}

		logrus.SetFormatter(&logrus.JSONFormatter{})
		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			"git@github.com:freshteapot/test-deploy-key.git",
			"/tmp/dolittle-k8s",
			gitRepoBranch,
			// TODO fix this, then update deployment
			"/Users/freshteapot/dolittle/.ssh/test-deploy",
		)

		ctx := context.TODO()
		kubeconfig, _ := cmd.Flags().GetString("kube-config")
		// TODO hoist localhost into viper
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println("todo", ctx, client, gitRepo)

		pathToFile := args[0]
		b, err := ioutil.ReadFile(pathToFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		parts := strings.Split(string(b), `---`)
		for _, part := range parts {
			if part == "" {
				continue
			}
			fmt.Println("# Before")
			err = doSSA([]byte(part), ctx, config)
			fmt.Println(err)
			fmt.Println("# After")
		}
	},
}

func init() {
	RootCmd.AddCommand(buildNatsCMD)
	buildNatsCMD.Flags().String("kube-config", "", "FullPath to kubeconfig")
}

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func doSSA(body []byte, ctx context.Context, cfg *rest.Config) error {

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(body, nil, obj)
	if err != nil {
		return err
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	obj.SetNamespace("application-11b6cf47-5d9f-438f-8116-0d9828654657")
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels["test"] = "nats11"
	obj.SetLabels(labels)

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations["test"] = "nats22"
	obj.SetAnnotations(annotations)

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())

	//if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
	//	// namespaced resources should specify the namespace
	//	dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	//} else {
	//	// for cluster-wide resources
	//	dr = dyn.Resource(mapping.Resource)
	//}

	// 6. Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	fmt.Println(string(data), dr)
	//return nil
	// 7. Create or Update the object with SSA
	//     types.ApplyPatchType indicates SSA.
	//     FieldManager specifies the field owner ID.
	//_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
	//	FieldManager: "platform-api",
	//})

	err = dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})

	return err
}
