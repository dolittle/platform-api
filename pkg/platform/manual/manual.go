package manual

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Repo struct {
	client     kubernetes.Interface
	k8sRepoV2  k8s.Repo
	logContext logrus.FieldLogger
}

func NewManualHelper(
	client kubernetes.Interface,
	k8sRepoV2 k8s.Repo,
	logContext logrus.FieldLogger,
) Repo {
	return Repo{
		client:     client,
		k8sRepoV2:  k8sRepoV2,
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
