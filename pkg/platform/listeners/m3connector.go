package listeners

// Inspired by https://github.com/feiskyer/kubernetes-handbook/blob/master/examples/client/informer/informer.go
// Inspired by https://github.com/heptiolabs/eventrouter/blob/master/main.go

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type m3ConnectorController struct {
	informerFactory   informers.SharedInformerFactory
	configMapInformer coreinformers.ConfigMapInformer
	logContext        logrus.FieldLogger
	gitSync           gitStorage.GitSync
	repo              storage.RepoApplication
}

func (c *m3ConnectorController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.configMapInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *m3ConnectorController) getEnvironment(resource *corev1.ConfigMap) (storage.JSONEnvironment, error) {
	customerID := resource.Annotations["dolittle.io/tenant-id"]
	applicationID := resource.Annotations["dolittle.io/application-id"]

	application, err := c.repo.GetApplication(customerID, applicationID)
	if err != nil {
		// This can be noisy due to the plaftform environment :(
		if !errors.Is(err, storage.ErrNotFound) {
			c.logContext.WithField("error", err).Error("error loading GetApplication")
		}
		return storage.JSONEnvironment{}, storage.ErrNotFound
	}

	// Find environment
	found := funk.Find(application.Environments, func(environment storage.JSONEnvironment) bool {
		return environment.Name == resource.Labels["environment"]
	})

	if found == nil {
		c.logContext.WithField("environment", resource.Labels["environment"]).Error("environment not found")
		return storage.JSONEnvironment{}, storage.ErrNotFound
	}

	return found.(storage.JSONEnvironment), nil
}

func (c *m3ConnectorController) saveEnvironment(resource *corev1.ConfigMap, environment storage.JSONEnvironment) error {
	customerID := resource.Annotations["dolittle.io/tenant-id"]
	applicationID := resource.Annotations["dolittle.io/application-id"]

	application, err := c.repo.GetApplication(customerID, applicationID)
	if err != nil {
		return err
	}

	for index, currentEnvironment := range application.Environments {
		if currentEnvironment.Name != environment.Name {
			continue
		}
		application.Environments[index] = environment
	}

	return c.repo.SaveApplication(application)
}

func (c *m3ConnectorController) upsert(resource *corev1.ConfigMap) {
	environment, err := c.getEnvironment(resource)
	if err != nil {
		// This can be noisy due to the plaftform environment :(
		if !errors.Is(err, storage.ErrNotFound) {
			c.logContext.WithField("error", err).Error("error loading GetApplication")
		}
		return
	}

	if environment.Connections.M3Connector {
		return
	}

	environment.Connections.Kafka = true
	environment.Connections.M3Connector = true

	err = c.saveEnvironment(resource, environment)
	if err != nil {
		c.logContext.WithField("error", err).Error("failed to save environment")
	}

	c.logContext.Info("application is m3connector aware")
}

func (c *m3ConnectorController) add(obj interface{}) {
	resource := obj.(*corev1.ConfigMap)
	c.upsert(resource)
}

func (c *m3ConnectorController) update(old, new interface{}) {
	resource := new.(*corev1.ConfigMap)
	c.upsert(resource)
}

func (c *m3ConnectorController) delete(obj interface{}) {
	resource := obj.(*corev1.ConfigMap)
	c.logContext.Infof("DELETED: %s %s/%s", resource.APIVersion, resource.Namespace, resource.Name)

	environment, err := c.getEnvironment(resource)
	if err != nil {
		// This can be noisy due to the plaftform environment :(
		if !errors.Is(err, storage.ErrNotFound) {
			c.logContext.WithField("error", err).Error("error loading GetApplication")
		}
		return
	}

	environment.Connections.Kafka = false
	environment.Connections.M3Connector = false
	err = c.saveEnvironment(resource, environment)
	if err != nil {
		c.logContext.WithField("error", err).Error("failed to save environment")
	}

	c.logContext.Info("application updated, m3connector no longer enabled for this environment")

}

func NewM3ConnectorConfigmapListenerController(
	informerFactory informers.SharedInformerFactory,
	gitSync gitStorage.GitSync,
	repo storage.RepoApplication,
	logContext logrus.FieldLogger,
) *m3ConnectorController {
	configMapInformer := informerFactory.Core().V1().ConfigMaps()

	c := &m3ConnectorController{
		informerFactory:   informerFactory,
		configMapInformer: configMapInformer,
		logContext:        logContext,
		gitSync:           gitSync,
		repo:              repo,
	}

	// FilteringResourceEventHandler
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			resource := obj.(*corev1.ConfigMap)

			if !strings.HasPrefix(resource.Namespace, "application-") {
				return false
			}

			if !strings.HasSuffix(resource.Name, "-kafka-files") {
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
	configMapInformer.Informer().AddEventHandler(handler)
	return c
}

func NewM3ConnectorConfigmapListener(
	client kubernetes.Interface,
	gitSync gitStorage.GitSync,
	repo storage.RepoApplication,
	logContext logrus.FieldLogger,
) {
	// TODO do I need a name space?
	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Hour*24)
	controller := NewM3ConnectorConfigmapListenerController(factory, gitSync, repo, logContext)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		logContext.Fatal(err)
	}
	select {}
}
