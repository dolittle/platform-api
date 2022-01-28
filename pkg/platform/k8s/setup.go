package k8s

import (
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func InitKubernetesClient() (kubernetes.Interface, *rest.Config) {
	kubeconfig := viper.GetString("tools.server.kubeConfig")

	if kubeconfig == "incluster" {
		kubeconfig = ""
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return client, config
}
