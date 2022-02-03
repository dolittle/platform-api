package manual

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Repo struct {
	client     kubernetes.Interface
	k8sRepo    platformK8s.K8sRepo
	logContext logrus.FieldLogger
}

func NewManualHelper(client kubernetes.Interface, k8sRepo platformK8s.K8sRepo, logContext logrus.FieldLogger) Repo {
	return Repo{
		client:     client,
		k8sRepo:    k8sRepo,
		logContext: logContext,
	}
}

func (r Repo) GatherOne(namespace string) {
	//application := storage.JSONApplication2{}
	ctx := context.TODO()
	client := r.client

	namespaceResource, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err != nil {
		panic(err)
	}

	if !automate.IsApplicationNamespace(*namespaceResource) {
		r.logContext.WithFields(logrus.Fields{
			"namespace": namespace,
		}).Info("Namespace not dolittle application")
		return
	}

	//Get customerTenants
	// TODO write this to storage?
	items := r.GetCustomerTenants(ctx, namespace)
	// Get Environments

	b, _ := json.Marshal(items)
	fmt.Println(string(b))
	//fmt.Println(applications)
}
