package automate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// SetRuntimeObjectGVK set's the given objects GroupVersionKind to the one the schema
// is aware of. This is because the k8s API doesn't always return the objects correct
// TypeMeta with these fields populated.
// See https://github.com/kubernetes/kubernetes/issues/3030#issuecomment-700099699
func SetRuntimeObjectGVK(schema *runtime.Scheme, runtimeObject runtime.Object) error {
	// get the GroupVersionKind of the object type from the schema
	gvks, _, err := schema.ObjectKinds(runtimeObject)
	if err != nil {
		return err
	}
	// set the runtimeObjects GroupVersionKind to match the one that the schema knows of
	runtimeObject.GetObjectKind().SetGroupVersionKind(gvks[0])
	return nil
}

// WriteResourceToFile serializes and writes the given resource to the given directory and file
// The resource should already have it's ManagedFields and ResourceVersion's cleared out to avoid noise in the file
func WriteResourceToFile(microserviceDirectory string, fileName string, resource runtime.Object, serializer *k8sJson.Serializer) error {
	err := os.MkdirAll(microserviceDirectory, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(microserviceDirectory, fileName))
	if err != nil {
		return err
	}

	defer file.Close()
	err = serializer.Encode(resource, file)
	if err != nil {
		return err
	}

	return nil
}

func InitializeSchemeAndSerializerForConfigMap() (*runtime.Scheme, *k8sJson.Serializer, error) {
	// based of https://github.com/kubernetes/kubernetes/issues/3030#issuecomment-700099699
	// create the scheme and make it aware of the corev1 & appv1 types
	scheme := runtime.NewScheme()
	err := corev1.AddToScheme(scheme)
	if err != nil {
		return scheme, nil, err
	}
	err = appsv1.AddToScheme(scheme)
	if err != nil {
		return scheme, nil, err
	}

	serializer := k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory,
		scheme,
		scheme,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
	return scheme, serializer, nil
}

func SerializeRuntimeObject(runtimeObject runtime.Object) []byte {
	scheme, serializer, err := InitializeSchemeAndSerializerForConfigMap()
	if err != nil {
		panic(err.Error())
	}

	SetRuntimeObjectGVK(scheme, runtimeObject)

	var buf bytes.Buffer
	_ = serializer.Encode(runtimeObject, &buf)
	return buf.Bytes()
}

func GetDolittleConfigMaps(ctx context.Context, client kubernetes.Interface, namespace string) ([]corev1.ConfigMap, error) {
	results := make([]corev1.ConfigMap, 0)
	configmaps, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return results, err
	}

	for _, configMap := range configmaps.Items {
		if !strings.HasSuffix(configMap.GetName(), "-dolittle") {
			continue
		}
		results = append(results, configMap)
	}
	return results, nil
}

func GetDolittleConfigMap(ctx context.Context, client kubernetes.Interface, applicationID string, environment string, microserviceID string) (*corev1.ConfigMap, error) {
	namespace := fmt.Sprintf("application-%s", applicationID)
	var result *corev1.ConfigMap
	configmaps, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return result, err
	}

	for _, configMap := range configmaps.Items {
		labels := configMap.GetLabels()
		annotations := configMap.GetAnnotations()

		if annotations["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		if labels["environment"] != environment {
			continue
		}

		if !strings.HasSuffix(configMap.GetName(), "-dolittle") {
			continue
		}

		return &configMap, nil
	}
	return result, errors.New("not.found")
}

func GetDeployments(ctx context.Context, client kubernetes.Interface, namespace string) ([]appsv1.Deployment, error) {
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var microserviceDeployments []appsv1.Deployment
	for _, deployment := range deployments.Items {
		if _, ok := deployment.Labels["microservice"]; !ok {
			continue
		}
		if _, ok := deployment.Annotations["dolittle.io/microservice-id"]; !ok {
			continue
		}
		microserviceDeployments = append(microserviceDeployments, deployment)
	}
	return microserviceDeployments, nil
}

// GetRuntimeDeployment get's a microservices deployment and the index of the Runtime container within that deployment
func GetRuntimeDeployment(ctx context.Context, client kubernetes.Interface, applicationID string, environment string, microserviceID string) (int, appsv1.Deployment, error) {
	namespace := fmt.Sprintf("application-%s", applicationID)

	deployments, err := GetDeployments(ctx, client, namespace)
	if err != nil {
		return -1, appsv1.Deployment{}, err
	}

	for _, deployment := range deployments {
		if deployment.Annotations["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		if deployment.Labels["environment"] != environment {
			continue
		}

		for index, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == "runtime" {
				return index, deployment, nil
			}
		}
	}

	return -1, appsv1.Deployment{}, errors.New("not.found")
}

func ConvertObjectMetaToMicroservice(objectMeta metav1.Object) configK8s.Microservice {
	labels := objectMeta.GetLabels()
	annotations := objectMeta.GetAnnotations()

	microserviceID := annotations["dolittle.io/microservice-id"]
	microserviceName := labels["microservice"]
	customerTenant := configK8s.Tenant{
		Name: labels["tenant"],
		ID:   annotations["dolittle.io/tenant-id"],
	}
	k8sApplication := configK8s.Application{
		Name: labels["application"],
		ID:   annotations["dolittle.io/application-id"],
	}

	environment := labels["environment"]

	var kind platform.MicroserviceKind
	if kindString, ok := annotations["dolittle.io/microservice-kind"]; !ok {
		kind = platform.MicroserviceKindUnknown
	} else {
		kind = platform.MicroserviceKind(kindString)
	}

	return configK8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customerTenant,
		Application: k8sApplication,
		Environment: environment,
		ResourceID:  "",
		Kind:        kind,
	}
}

func GetAllCustomerMicroservices(ctx context.Context, client kubernetes.Interface) ([]configK8s.Microservice, error) {
	microservices := make([]configK8s.Microservice, 0)
	deployments := make([]appsv1.Deployment, 0)
	namespaces := GetNamespaces(ctx, client)
	for _, namespace := range namespaces {
		if !IsApplicationNamespace(namespace) {
			continue
		}
		specific, err := GetDeployments(ctx, client, namespace.Name)
		if err != nil {
			return microservices, err
		}
		deployments = append(deployments, specific...)
	}

	for _, deployment := range deployments {
		microservice := ConvertObjectMetaToMicroservice(deployment.GetObjectMeta())
		microservices = append(microservices, microservice)
	}
	return microservices, nil
}

func GetNamespaces(ctx context.Context, client kubernetes.Interface) []corev1.Namespace {
	namespacesList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return namespacesList.Items
}

func IsApplicationNamespace(namespace corev1.Namespace) bool {
	if !strings.HasPrefix(namespace.GetName(), "application-") {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := namespace.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := namespace.Labels["application"]; !hasLabel {
		return false
	}

	return true
}

func GetMicroserviceDirectory(rootFolder string, objectMeta metav1.Object) string {
	labels := objectMeta.GetLabels()
	customer := labels["tenant"]
	application := labels["application"]
	environment := labels["environment"]
	microservice := labels["microservice"]

	return filepath.Join(rootFolder,
		"Source",
		"V3",
		"Kubernetes",
		"Customers",
		customer,
		application,
		environment,
		microservice,
	)
}

func AddDataToConfigMap(ctx context.Context, client kubernetes.Interface, logContext logrus.FieldLogger, key string, data []byte, configMap corev1.ConfigMap, dryRun bool) error {
	microservice := ConvertObjectMetaToMicroservice(configMap.GetObjectMeta())
	logContext.WithFields(logrus.Fields{
		"microservice_id": microservice.ID,
		"microservice":    microservice.Name,
		"application_id":  microservice.Application.ID,
		"environment":     microservice.Environment,
		"namespace":       configMap.Namespace,
		"function":        "AddDataToConfigMap",
	})

	if configMap.Data == nil {
		// TODO this is a sign it might not be using a runtime, maybe we skip
		configMap.Data = make(map[string]string)
	}

	if _, ok := configMap.Data[key]; ok {
		logContext.Infof("Key %s already exists in the configmap, skipping the update", key)
		return nil
	}

	configMap.Data[key] = string(data)

	if dryRun {
		bytes := SerializeRuntimeObject(&configMap)

		logContext = logContext.WithField("data", string(bytes))
		logContext.Info("--dry-run enabled, skipping update")
		return nil
	}

	_, err := client.CoreV1().ConfigMaps(configMap.Namespace).Update(ctx, &configMap, metav1.UpdateOptions{})
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("updating configmap")
		return errors.New("update.failed")
	}
	return nil
}

// AddVolumeMountToContainer adds the given volumeMount to the container specified by containerIndex within the given deployment
func AddVolumeMountToContainer(ctx context.Context,
	client kubernetes.Interface,
	logContext logrus.FieldLogger,
	volumeMount corev1.VolumeMount,
	containerIndex int,
	deployment appsv1.Deployment,
	dryRun bool) error {
	microservice := ConvertObjectMetaToMicroservice(deployment.GetObjectMeta())

	logContext.WithFields(logrus.Fields{
		"microservice_id": microservice.ID,
		"microservice":    microservice.Name,
		"application_id":  microservice.Application.ID,
		"application":     microservice.Application.Name,
		"environment":     microservice.Environment,
		"namespace":       deployment.Namespace,
		"function":        "AddVolumeMountToRuntimeContainer",
		"mount_path":      volumeMount.MountPath,
		"sub_path":        volumeMount.SubPath,
	})

	container := deployment.Spec.Template.Spec.Containers[containerIndex]

	hasPlatformMount := false
	for _, containerMounts := range container.VolumeMounts {
		if containerMounts.SubPath == volumeMount.SubPath {
			hasPlatformMount = true
		}
	}

	if hasPlatformMount {
		logContext.Infof("Container already had a volumemount on subpath %s", volumeMount.SubPath)
		return nil
	}

	container.VolumeMounts = append(container.VolumeMounts, volumeMount)
	// mutate the deployment with the modified container
	deployment.Spec.Template.Spec.Containers[containerIndex] = container

	if dryRun {
		bytes := SerializeRuntimeObject(&deployment)
		logContext.WithFields(logrus.Fields{
			"data": string(bytes),
		}).Info("--dry-run enabled, skipping the update")
		return nil
	}

	_, err := client.AppsV1().Deployments(deployment.Namespace).Update(ctx, &deployment, metav1.UpdateOptions{})
	if err != nil {
		logContext.Fatal(err.Error())
	}

	return nil
}
