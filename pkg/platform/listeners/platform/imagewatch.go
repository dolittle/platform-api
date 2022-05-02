package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	ErrNotFound = errors.New("not-found")
)

type imageWatchController struct {
	informerFactory informers.SharedInformerFactory
	informer        appsinformers.DeploymentInformer
	logContext      logrus.FieldLogger
	repo            storage.RepoMicroservice
}

func (c *imageWatchController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.informer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *imageWatchController) hasCorrectAnnotationsOrLabelsMissing(resource *appsv1.Deployment) bool {
	logContext := c.logContext.WithField("namespace", resource.Namespace)

	if _, ok := resource.Annotations["dolittle.io/tenant-id"]; !ok {
		logContext.WithFields(logrus.Fields{
			"missing":      "annotations",
			"key":          "dolittle.io/tenant-id",
			"resourceName": resource.Name,
		}).Warn("missing annotation")
		return false
	}

	if _, ok := resource.Annotations["dolittle.io/application-id"]; !ok {
		logContext.WithFields(logrus.Fields{
			"missing":      "annotations",
			"key":          "dolittle.io/application-id",
			"resourceName": resource.Name,
		}).Warn("missing annotation")
		return false
	}

	if _, ok := resource.Annotations["dolittle.io/microservice-id"]; !ok {
		logContext.WithFields(logrus.Fields{
			"missing":      "annotations",
			"key":          "dolittle.io/microservice-id",
			"resourceName": resource.Name,
		}).Warn("missing annotation")
		return false
	}

	if _, ok := resource.Labels["environment"]; !ok {
		logContext.WithFields(logrus.Fields{
			"missing":      "label",
			"key":          "environment",
			"resourceName": resource.Name,
		}).Warn("missing label")
		return false
	}

	return true
}

func (c *imageWatchController) upsert(resource *appsv1.Deployment) {
	if !c.hasCorrectAnnotationsOrLabelsMissing(resource) {
		return
	}

	customerID := resource.Annotations["dolittle.io/tenant-id"]
	applicationID := resource.Annotations["dolittle.io/application-id"]
	environmnet := resource.Labels["environment"]
	microserviceID := resource.Annotations["dolittle.io/microservice-id"]

	logContext := c.logContext.WithFields(logrus.Fields{
		"customer_id":     customerID,
		"application_id":  applicationID,
		"environment":     environmnet,
		"microservice_id": microserviceID,
	})

	raw, err := c.repo.GetMicroservice(customerID, applicationID, environmnet, microserviceID)
	if err != nil {
		// This can be noisy due to the platform environment :(, if we were to move
		// to dev cluster for dev, this becomes less noisy :)
		if !errors.Is(err, storage.ErrNotFound) {
			// TODO I wonder if this should be ErrNotFound in GetMicroservice
			// TODO this doesn't work well as we don't have current microservices in the storage
			if !strings.Contains(err.Error(), "no such file or directory") {
				logContext.WithField("error", err).Error("error getting microservice info")
			}

		}
		return
	}

	var microserviceBase platform.HttpMicroserviceBase
	err = json.Unmarshal(raw, &microserviceBase)
	if err != nil {
		logContext.WithField("error", err).Error("error internal json")
		return
	}

	if microserviceBase.Kind != platform.MicroserviceKindSimple {
		logContext.WithField("error", microserviceBase.Kind).Error("kind not supported")
	}

	var ms platform.HttpInputSimpleInfo
	err = json.Unmarshal(raw, &ms)
	if err != nil {
		logContext.WithField("error", err).Error("error internal json simple microservice ")
		return
	}

	// Check head

	headImage, err := c.getHeadImage(resource.Spec.Template.Spec.Containers)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			logContext.WithField("error", err).Error("error c.getHeadImage")
		}
		return
	}

	fromChanges := make(map[string]string)
	toChanges := make(map[string]string)
	update := false
	if ms.Extra.Headimage != headImage {
		fromChanges["headImage"] = ms.Extra.Headimage
		toChanges["headImage"] = headImage
		ms.Extra.Headimage = headImage
		update = true
	}

	// Save progress
	if !update {
		return
	}

	logContext.WithFields(logrus.Fields{
		"changeset": logrus.Fields{
			"from": fromChanges,
			"to":   toChanges,
		},
	}).Info("updated deployment")
}

func (c *imageWatchController) getHeadImage(containers []corev1.Container) (string, error) {
	filtered := funk.Filter(containers, func(container corev1.Container) bool {
		return container.Name == "head"
	}).([]corev1.Container)

	if len(filtered) == 0 {
		return "", ErrNotFound
	}

	return filtered[0].Image, nil
}

func (c *imageWatchController) add(obj interface{}) {
	resource := obj.(*appsv1.Deployment)
	c.upsert(resource)
}

func (c *imageWatchController) update(old, new interface{}) {
	resource := new.(*appsv1.Deployment)
	c.upsert(resource)
}

func (c *imageWatchController) delete(obj interface{}) {
	resource := obj.(*appsv1.Deployment)
	fmt.Println(resource.Namespace, resource.Name)
	//environment, err := c.getEnvironment(resource)
	//if err != nil {
	//	// This can be noisy due to the platform environment :(, if we were to move
	//	// to dev cluster for dev, this becomes less noisy :)
	//	if !errors.Is(err, storage.ErrNotFound) {
	//		c.logContext.WithField("error", err).Error("error getting environment info")
	//	}
	//	return
	//}
	//
	//environment.Connections.M3Connector = false
	//err = c.saveEnvironment(resource, environment)
	//if err != nil {
	//	c.logContext.WithField("error", err).Error("failed to save environment")
	//	return
	//}
	//
	//c.logContext.Info("application updated, m3connector no longer enabled for this environment")
}

func NewImageWatchListenerController(
	informerFactory informers.SharedInformerFactory,
	repo storage.RepoMicroservice,
	logContext logrus.FieldLogger,
) *imageWatchController {
	informer := informerFactory.Apps().V1().Deployments()

	c := &imageWatchController{
		informerFactory: informerFactory,
		informer:        informer,
		logContext:      logContext,
		repo:            repo,
	}

	// FilteringResourceEventHandler
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			resource := obj.(*appsv1.Deployment)
			// Only focus on application namespaces
			if !strings.HasPrefix(resource.Namespace, "application-") {
				return false
			}

			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    c.add,
			UpdateFunc: c.update,
			DeleteFunc: c.delete,
		},
	}
	informer.Informer().AddEventHandler(handler)
	return c
}

// NewImageWatchListener listen to deployment changes
func NewImageWatchListener(
	client kubernetes.Interface,
	repo storage.RepoMicroservice,
	logContext logrus.FieldLogger,
) {
	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Hour*24)
	controller := NewImageWatchListenerController(factory, repo, logContext)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		logContext.Fatal(err)
	}
	select {}
}
