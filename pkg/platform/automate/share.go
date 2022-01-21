package automate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func SetConfigMapGVK(schema *runtime.Scheme, configMap *corev1.ConfigMap) error {
	// get the GroupVersionKind of the configMap type from the schema
	gvks, _, err := schema.ObjectKinds(configMap)
	if err != nil {
		return err
	}
	// set the configMaps GroupVersionKind to match the one that the schema knows of
	configMap.GetObjectKind().SetGroupVersionKind(gvks[0])
	return nil
}

func InitializeSchemeAndSerializerForConfigMap() (*runtime.Scheme, *k8sJson.Serializer, error) {
	// based of https://github.com/kubernetes/kubernetes/issues/3030#issuecomment-700099699
	// create the scheme and make it aware of the corev1 types
	scheme := runtime.NewScheme()
	err := corev1.AddToScheme(scheme)
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

func DumpConfigMap(configMap *corev1.ConfigMap) []byte {
	scheme, serializer, err := InitializeSchemeAndSerializerForConfigMap()
	if err != nil {
		panic(err.Error())
	}

	SetConfigMapGVK(scheme, configMap)

	var buf bytes.Buffer

	_ = serializer.Encode(configMap, &buf)
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

func GetOneDolittleConfigMap(ctx context.Context, client kubernetes.Interface, applicationID string, environment string, microserviceID string) (*corev1.ConfigMap, error) {
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

	return configK8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customerTenant,
		Application: k8sApplication,
		Environment: environment,
		// TODO wwhen we explicitly set the kind as an annotation we can rely on it
		ResourceID: "",
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

func GetMicroserviceDirectory(rootFolder string, configMap corev1.ConfigMap) string {
	labels := configMap.GetObjectMeta().GetLabels()
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

func UpdateConfigMap(ctx context.Context, client kubernetes.Interface, logContext logrus.FieldLogger, configMap corev1.ConfigMap, dryRun bool) error {
	microservice := ConvertObjectMetaToMicroservice(configMap.GetObjectMeta())
	platform := configK8s.NewMicroserviceConfigmapPlatformData(microservice)
	b, _ := json.MarshalIndent(platform, "", "  ")
	platformJSON := string(b)

	if configMap.Data == nil {
		// TODO this is a sign it might not be using a runtime, maybe we skip
		configMap.Data = make(map[string]string)
	}

	configMap.Data["platform.json"] = platformJSON

	namespace := configMap.Namespace

	logContext.WithFields(logrus.Fields{
		"microservice_id": microservice.ID,
		"application_id":  microservice.Application.ID,
		"environment":     microservice.Environment,
		"namespace":       microservice.Environment,
	})

	if dryRun {
		b := DumpConfigMap(&configMap)

		logContext = logContext.WithField("data", string(b))
		logContext.Info("Would write")
		return nil
	}

	_, err := client.CoreV1().ConfigMaps(namespace).Update(ctx, &configMap, metav1.UpdateOptions{})
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("updating configmap")
		return errors.New("update.failed")
	}
	return nil
}
