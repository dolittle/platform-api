package m3connector

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type kafkaFilesController struct {
	informerFactory   informers.SharedInformerFactory
	configMapInformer coreinformers.ConfigMapInformer
	logContext        logrus.FieldLogger
	repo              storage.RepoApplication
}

func (c *kafkaFilesController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.configMapInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *kafkaFilesController) hasCorrectAnnotationsOrLabelsMissing(resource *corev1.ConfigMap) bool {
	logContext := c.logContext.WithField("namespace", resource.Namespace)

	if _, ok := resource.Annotations["dolittle.io/tenant-id"]; !ok {
		logContext.WithFields(logrus.Fields{
			"missing": "annotations",
			"key":     "dolittle.io/tenant-id",
		}).Warn("missing annotation")
		return false
	}

	if _, ok := resource.Annotations["dolittle.io/application-id"]; !ok {
		logContext.WithFields(logrus.Fields{
			"missing": "annotations",
			"key":     "dolittle.io/application-id",
		}).Warn("missing annotation")
		return false
	}

	if _, ok := resource.Labels["environment"]; !ok {
		logContext.Warn("missing label environment")
		logContext.WithFields(logrus.Fields{
			"missing": "label",
			"key":     "environment",
		}).Warn("missing label")
		return false
	}

	return true
}

func (c *kafkaFilesController) getEnvironment(resource *corev1.ConfigMap) (storage.JSONEnvironment, error) {
	customerID := resource.Annotations["dolittle.io/tenant-id"]
	applicationID := resource.Annotations["dolittle.io/application-id"]

	if customerID == "" || applicationID == "" {
		return storage.JSONEnvironment{}, storage.ErrNotFound
	}

	application, err := c.repo.GetApplication(customerID, applicationID)
	if err != nil {
		return storage.JSONEnvironment{}, err
	}

	environment, err := storage.GetEnvironment(application.Environments, resource.Labels["environment"])

	if err != nil {
		return storage.JSONEnvironment{}, err
	}

	return environment, nil
}

func (c *kafkaFilesController) saveEnvironment(resource *corev1.ConfigMap, environment storage.JSONEnvironment) error {
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

func (c *kafkaFilesController) upsert(resource *corev1.ConfigMap) {
	if !c.hasCorrectAnnotationsOrLabelsMissing(resource) {
		return
	}

	// TODO this should be revisited when we look at rebuilding an empty cluster
	// Having the source of truth, mixed with listening for changes will need more logic
	environment, err := c.getEnvironment(resource)
	if err != nil {
		// This can be noisy due to the platform environment :(, if we were to move
		// to dev cluster for dev, this becomes less noisy :)
		if !errors.Is(err, storage.ErrNotFound) {
			c.logContext.WithField("error", err).Error("error getting environment info")
		}
		return
	}

	if environment.Connections.M3Connector {
		return
	}

	environment.Connections.M3Connector = true

	err = c.saveEnvironment(resource, environment)
	if err != nil {
		c.logContext.WithField("error", err).Error("failed to save environment")
		return
	}

	c.logContext.Info("application is m3connector aware")
}

func (c *kafkaFilesController) add(obj interface{}) {
	resource := obj.(*corev1.ConfigMap)
	c.upsert(resource)
}

func (c *kafkaFilesController) update(old, new interface{}) {
	resource := new.(*corev1.ConfigMap)
	c.upsert(resource)
}

func (c *kafkaFilesController) delete(obj interface{}) {
	resource := obj.(*corev1.ConfigMap)

	environment, err := c.getEnvironment(resource)
	if err != nil {
		// This can be noisy due to the platform environment :(, if we were to move
		// to dev cluster for dev, this becomes less noisy :)
		if !errors.Is(err, storage.ErrNotFound) {
			c.logContext.WithField("error", err).Error("error getting environment info")
		}
		return
	}

	environment.Connections.M3Connector = false
	err = c.saveEnvironment(resource, environment)
	if err != nil {
		c.logContext.WithField("error", err).Error("failed to save environment")
		return
	}

	c.logContext.Info("application updated, m3connector no longer enabled for this environment")
}

func NewKafkaFilesConfigmapListenerController(
	informerFactory informers.SharedInformerFactory,
	repo storage.RepoApplication,
	logContext logrus.FieldLogger,
) *kafkaFilesController {
	configMapInformer := informerFactory.Core().V1().ConfigMaps()

	c := &kafkaFilesController{
		informerFactory:   informerFactory,
		configMapInformer: configMapInformer,
		logContext:        logContext,
		repo:              repo,
	}

	// FilteringResourceEventHandler
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			resource := obj.(*corev1.ConfigMap)
			// Only focus on application namespaces
			if !strings.HasPrefix(resource.Namespace, "application-") {
				return false
			}

			// Only focus on the file linked to m3connector
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

func NewKafkaFilesConfigmapListener(
	client kubernetes.Interface,
	repo storage.RepoApplication,
	logContext logrus.FieldLogger,
) {
	// TODO do I need a name space?
	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Hour*24)
	controller := NewKafkaFilesConfigmapListenerController(factory, repo, logContext)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		logContext.Fatal(err)
	}
	select {}
}
