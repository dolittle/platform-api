package job

// Inspired by https://github.com/feiskyer/kubernetes-handbook/blob/master/examples/client/informer/informer.go
// Inspired by https://github.com/heptiolabs/eventrouter/blob/master/main.go

import (
	"fmt"
	"strings"
	"time"

	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type customerController struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
	logContext      logrus.FieldLogger
	gitSync         gitStorage.GitSync
}

func (c *customerController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *customerController) podAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	// I think I just watched this explode because startTime was absent
	startTime := "n/a"
	if pod.Status.StartTime != nil {
		pod.Status.StartTime.UTC().Format(time.RFC3339)
	}
	c.logContext.Infof("POD CREATED: %s/%s %s", pod.Namespace, pod.Name, startTime)
}

func (c *customerController) podUpdate(old, new interface{}) {
	oldPod := old.(*corev1.Pod)
	newPod := new.(*corev1.Pod)
	// TODO why didn't this log?
	c.logContext.Infof(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)

	for _, status := range newPod.Status.InitContainerStatuses {
		// TODO we can warn in teams that there is a problem
		cmd := fmt.Sprintf(`kubectl -n %s logs %s -c %s`,
			newPod.Namespace,
			newPod.Name,
			status.Name,
		)
		fmt.Println(cmd)

		if status.State.Running != nil {
			fmt.Println("Running", status.State.Running.StartedAt)
		}

		if status.State.Waiting != nil {
			fmt.Println("Waiting", status.State.Waiting.Message, status.State.Waiting.Reason)
		}

		if status.State.Terminated != nil {
			fmt.Println("Terminated", status.State.Terminated.Message, status.State.Terminated.Reason)
		}

		fmt.Println("")
	}

	if newPod.Status.Phase == "Succeeded" {
		// trigger gitPull
		err := c.gitSync.Pull()
		if err != nil {
			c.logContext.WithFields(logrus.Fields{
				"error":   err,
				"context": "customer-created-update-repo",
			}).Fatal("Failed to update repo")
		}

		customerID := newPod.Annotations["dolittle.io/tenant-id"]

		c.logContext.WithFields(logrus.Fields{
			"customer_id": customerID,
			"context":     "customer-created-update-repo",
		}).Info("Repo updated with changes after job successfully ran")
	}
}

func (c *customerController) podDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)
	c.logContext.Infof("POD DELETED: %s/%s", pod.Namespace, pod.Name)
}

func NewCustomerListenerController(informerFactory informers.SharedInformerFactory, gitSync gitStorage.GitSync, logContext logrus.FieldLogger) *customerController {
	podInformer := informerFactory.Core().V1().Pods()

	c := &customerController{
		informerFactory: informerFactory,
		podInformer:     podInformer,
		logContext:      logContext,
		gitSync:         gitSync,
	}

	// FilteringResourceEventHandler
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod := obj.(*corev1.Pod)
			if _, ok := pod.Labels["job-name"]; !ok {
				return false
			}

			if !strings.HasPrefix(pod.Name, "create-customer-") {
				return false
			}

			// Doing this here, so I can rely on it above
			if _, ok := pod.Annotations["dolittle.io/tenant-id"]; !ok {
				return false
			}

			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    c.podAdd,
			UpdateFunc: c.podUpdate,
			DeleteFunc: c.podDelete,
		},
	}
	podInformer.Informer().AddEventHandler(handler)
	return c
}

func NewCustomerJobListener(client kubernetes.Interface, gitSync gitStorage.GitSync, logContext logrus.FieldLogger) {
	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Hour*24, informers.WithNamespace("system-api"))
	controller := NewCustomerListenerController(factory, gitSync, logContext)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		logContext.Fatal(err)
	}
	select {}
}
