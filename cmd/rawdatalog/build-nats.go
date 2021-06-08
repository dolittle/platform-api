package rawdatalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

var buildNatsCMD = &cobra.Command{
	Use:   "build-nats",
	Short: "Building nats",
	Long: `
	go run main.go microservice build-nats --kube-config="/Users/freshteapot/.kube/config" --action=delete ./k8s/single-server-nats.yml
	`,
	Run: func(cmd *cobra.Command, args []string) {

		logrus.SetFormatter(&logrus.JSONFormatter{})

		ctx := context.TODO()
		kubeconfig, _ := cmd.Flags().GetString("kube-config")
		action, _ := cmd.Flags().GetString("action")

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		pathToFile := args[0]
		b, err := ioutil.ReadFile(pathToFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		namespace := "application-11b6cf47-5d9f-438f-8116-0d9828654657"

		parts := strings.Split(string(b), `---`)
		for _, part := range parts {
			if part == "" {
				continue
			}
			fmt.Println("# Before")
			err = doSSA(action, namespace, []byte(part), ctx, config)
			fmt.Println(err)
			fmt.Println("# After")
		}
	},
}

func init() {
	RootCmd.AddCommand(buildNatsCMD)
	buildNatsCMD.Flags().String("kube-config", "", "FullPath to kubeconfig")
	buildNatsCMD.Flags().String("action", "", "Action delete or upsert")
}

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func doSSA(action string, namespace string, body []byte, ctx context.Context, cfg *rest.Config) error {

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

	obj.SetNamespace(namespace)
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
	// Namespace specific
	dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())

	if action == "delete" {
		return dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	}

	if action != "upsert" {
		return errors.New("action not supported")
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "platform-api",
	})
	return err
}
