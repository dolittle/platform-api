package job

// Inspired by https://github.com/feiskyer/kubernetes-handbook/blob/master/examples/client/informer/informer.go
// Inspired by https://github.com/heptiolabs/eventrouter/blob/master/main.go

import (
	"fmt"
	"time"

	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// PodLoggingController logs the name and namespace of pods that are added,
// deleted, or updated
type PodLoggingController struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
	logContext      logrus.FieldLogger
	gitSync         gitStorage.GitSync
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *PodLoggingController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *PodLoggingController) podAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	// I think I just watched this explode because startTime was absent
	startTime := "n/a"
	if pod.Status.StartTime != nil {
		pod.Status.StartTime.UTC().Format(time.RFC3339)
	}
	c.logContext.Infof("POD CREATED: %s/%s %s", pod.Namespace, pod.Name, startTime)
}

func (c *PodLoggingController) podUpdate(old, new interface{}) {
	oldPod := old.(*corev1.Pod)
	newPod := new.(*corev1.Pod)
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

	// TODO might need logic to check initContainers is happy
	// TODO when job is finished (I think it will land here)
	if newPod.Status.Phase == "Succeeded" {
		// trigger gitPull
		err := c.gitSync.Pull()
		if err != nil {
			c.logContext.WithFields(logrus.Fields{
				"error":   err,
				"context": "application-created-update-repo",
			}).Fatal("Failed to update repo")
		}

		customerID := newPod.Annotations["dolittle.io/tenant-id"]
		applicationID := newPod.Annotations["dolittle.io/application-id"]

		//// TODO this is not needed
		//// TODO update state in the application
		//entry, err := c.repo.GetApplication2(customerID, applicationID)
		//if err != nil {
		//	c.logContext.WithFields(logrus.Fields{
		//		"error":   err,
		//		"context": "application-get-from-storage",
		//	}).Fatal("Failed to get data from storage")
		//}
		//
		//// TODO this should be a function that the job calls
		//entry.Status.State = storage.BuildStatusStateFinishedSuccess
		//entry.Status.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		//
		//err = c.repo.SaveApplication2(entry)
		//if err != nil {
		//	c.logContext.WithFields(logrus.Fields{
		//		"error":   err,
		//		"context": "application-save-to-storage",
		//	}).Fatal("Failed to save data to storage")
		//}

		c.logContext.WithFields(logrus.Fields{
			"customer_id":    customerID,
			"application_id": applicationID,
			"context":        "application-created-update-repo",
		}).Info("Repo updated with changes after job successfully ran")
	}
}

func (c *PodLoggingController) podDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)
	c.logContext.Infof("POD DELETED: %s/%s", pod.Namespace, pod.Name)
}

// NewPodLoggingController creates a PodLoggingController
func NewPodLoggingController(informerFactory informers.SharedInformerFactory, gitSync gitStorage.GitSync, logContext logrus.FieldLogger) *PodLoggingController {
	podInformer := informerFactory.Core().V1().Pods()

	c := &PodLoggingController{
		informerFactory: informerFactory,
		podInformer:     podInformer,
		logContext:      logContext,
		gitSync:         gitSync,
	}

	// FilteringResourceEventHandler
	a := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod := obj.(*corev1.Pod)
			if _, ok := pod.Labels["job-name"]; !ok {
				return false
			}

			// Doing this here, so I can rely on it above
			if _, ok := pod.Annotations["dolittle.io/tenant-id"]; !ok {
				return false
			}

			if _, ok := pod.Annotations["dolittle.io/application-id"]; !ok {
				return false
			}

			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: c.podAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.podUpdate,
			// Called on resource deletion.
			DeleteFunc: c.podDelete,
		},
	}
	podInformer.Informer().AddEventHandler(a)

	//podInformer.Informer().AddEventHandler(
	//	// Your custom resource event handlers.
	//	cache.ResourceEventHandlerFuncs{
	//		// Called on creation
	//		AddFunc: c.podAdd,
	//		// Called on resource update and every resyncPeriod on existing resources.
	//		UpdateFunc: c.podUpdate,
	//		// Called on resource deletion.
	//		DeleteFunc: c.podDelete,
	//	},
	//)
	return c
}

func NewListenForJobs(client kubernetes.Interface, gitSync gitStorage.GitSync, logContext logrus.FieldLogger) {
	// TODO What does resync do?
	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Hour*24, informers.WithNamespace("system-api"))
	controller := NewPodLoggingController(factory, gitSync, logContext)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		logContext.Fatal(err)
	}
	select {}
}
