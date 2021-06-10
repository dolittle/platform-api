package rawdatalog

import (
	_ "embed"
	"errors"
	"log"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

//go:embed k8s/single-server-nats.yml
var k8sRawDataLogIngestorNats string

//go:embed k8s/single-server-stan-memory.yml
var k8sRawDataLogIngestorStanInMemory string

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

type RawDataLogIngestorRepo struct {
	k8sClient       *kubernetes.Clientset
	k8sDolittleRepo platform.K8sRepo
}

func NewRawDataLogIngestorRepo(k8sDolittleRepo platform.K8sRepo, k8sClient *kubernetes.Clientset) RawDataLogIngestorRepo {
	return RawDataLogIngestorRepo{
		k8sClient:       k8sClient,
		k8sDolittleRepo: k8sDolittleRepo,
	}
}

func (r RawDataLogIngestorRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputRawDataLogIngestorInfo) error {
	config := r.k8sDolittleRepo.GetRestConfig()
	ctx := context.TODO()

	templates := []string{
		k8sRawDataLogIngestorNats,
		k8sRawDataLogIngestorStanInMemory,
	}

	labels := map[string]string{
		"tenant":       tenant.Name,
		"application":  application.Name,
		"environment":  input.Environment,
		"microservice": input.Name,
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":       tenant.ID,
		"dolittle.io/application-id":  application.ID,
		"dolittle.io/microservice-id": input.Dolittle.MicroserviceID,
	}

	// TODO changing writeTo will break this.
	if input.Extra.WriteTo != "stdout" {
		action := "upsert"
		for _, template := range templates {
			parts := strings.Split(template, `---`)
			for _, part := range parts {
				if part == "" {
					continue
				}

				err := doNats(
					labels, annotations,
					action, namespace, []byte(part), ctx, config)
				if err != nil {
					fmt.Println(err)
					return err
				}
			}
		}
	}

	// TODO add microservice
	err := r.doDolittle(namespace, tenant, application, applicationIngress, input)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (r RawDataLogIngestorRepo) Delete(namespace string, microserviceID string) error {
	// TODO This we might want to do different incase the files change
	//config := r.k8sDolittleRepo.GetRestConfig()
	//ctx := context.TODO()
	//templates := []string{
	//	k8sRawDataLogIngestorNats,
	//	k8sRawDataLogIngestorStanInMemory,
	//}

	client := r.k8sClient
	ctx := context.TODO()
	// Not possible to filter based on annotations
	opts := metav1.ListOptions{}
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, opts)

	if err != nil {
		return err
	}

	found := false
	// Ugly name
	var foundDeployment v1.Deployment
	for _, deployment := range deployments.Items {
		_, ok := deployment.ObjectMeta.Labels["microservice"]
		if !ok {
			continue
		}

		if deployment.ObjectMeta.Annotations["dolittle.io/microservice-id"] == microserviceID {
			found = true
			foundDeployment = deployment
			break
		}
	}

	if !found {
		return errors.New("not-found")
	}

	// Stop deployment
	s, err := client.AppsV1().
		Deployments(namespace).
		GetScale(ctx, foundDeployment.Name, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("issue")
	}

	sc := *s
	if sc.Spec.Replicas != 0 {
		sc.Spec.Replicas = 0
		_, err := client.AppsV1().
			Deployments(namespace).
			UpdateScale(ctx, foundDeployment.Name, &sc, metav1.UpdateOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Selector information for microservice, based on labels
	opts = metav1.ListOptions{
		LabelSelector: labels.FormatLabels(foundDeployment.GetObjectMeta().GetLabels()),
	}

	// Remove configmaps
	configs, _ := client.CoreV1().ConfigMaps(namespace).List(ctx, opts)

	for _, config := range configs.Items {
		err = client.CoreV1().ConfigMaps(namespace).Delete(ctx, config.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Remove secrets
	secrets, _ := client.CoreV1().Secrets(namespace).List(ctx, opts)
	for _, secret := range secrets.Items {
		err = client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Remove Ingress
	ingresses, _ := client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	for _, ingress := range ingresses.Items {
		err = client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove Network Policy
	policies, _ := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, opts)
	for _, policy := range policies.Items {
		err = client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, policy.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove Service
	services, _ := client.CoreV1().Services(namespace).List(ctx, opts)
	for _, service := range services.Items {
		err = client.CoreV1().Services(namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove statefulset
	statefulSets, _ := client.AppsV1().StatefulSets(namespace).List(ctx, opts)
	for _, stateful := range statefulSets.Items {
		err = client.AppsV1().StatefulSets(namespace).Delete(ctx, stateful.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove deployment
	err = client.AppsV1().
		Deployments(namespace).
		Delete(ctx, foundDeployment.Name, metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("todo")
	}

	return nil

	return errors.New("TODO")
}

func getKubernetesData(namespace string, body []byte, cfg *rest.Config) (*unstructured.Unstructured, dynamic.ResourceInterface, error) {

	var dr dynamic.ResourceInterface
	obj := &unstructured.Unstructured{}

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return obj, dr, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return obj, dr, err
	}

	_, gvk, err := decUnstructured.Decode(body, nil, obj)
	if err != nil {
		return obj, dr, err
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return obj, dr, err
	}

	obj.SetNamespace(namespace)

	dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	return obj, dr, nil
}

func doDo(obj *unstructured.Unstructured, dr dynamic.ResourceInterface, action string, ctx context.Context) error {
	//data, err := json.Marshal(obj)
	//fmt.Println(string(data))
	//return err
	if action == "delete" {
		return dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	}

	if action != "upsert" {
		return errors.New("action not supported")
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "platform-api",
	})
	return err
}

// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func doNats(
	labels map[string]string,
	annotations map[string]string,
	action string, namespace string, body []byte, ctx context.Context, cfg *rest.Config) error {
	obj, dr, err := getKubernetesData(namespace, body, cfg)
	if err != nil {
		return err
	}

	obj.SetNamespace(namespace)

	currentLabels := obj.GetLabels()
	if currentLabels == nil {
		currentLabels = map[string]string{}
	}

	mergedLabels := k8sLabels.Merge(currentLabels, labels)

	obj.SetLabels(mergedLabels)

	currentAnnotations := obj.GetAnnotations()
	if currentAnnotations == nil {
		currentAnnotations = map[string]string{}
	}

	mergedAnnotations := k8sLabels.Merge(currentAnnotations, annotations)
	obj.SetAnnotations(mergedAnnotations)

	return doDo(obj, dr, action, ctx)
}

func (r RawDataLogIngestorRepo) doDolittle(namespace string, tenant k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputRawDataLogIngestorInfo) error {

	// TODO not sure where this comes from really, assume dynamic
	customersTenantID := "17426336-fb8e-4425-8ab7-07d488367be9"

	environment := input.Environment
	host := applicationIngress.Host
	secretName := applicationIngress.SecretName

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      tenant,
		Application: application,
		Environment: environment,
		ResourceID:  customersTenantID,
	}

	ingressServiceName := strings.ToLower(fmt.Sprintf("%s-%s", microservice.Environment, microservice.Name))
	ingressRules := []k8s.SimpleIngressRule{
		{
			Path:            input.Extra.Ingress.Path,
			PathType:        networkingv1.PathType(input.Extra.Ingress.Pathtype),
			ServiceName:     ingressServiceName,
			ServicePortName: "http",
		},
	}

	// TODO do I need this?
	// TODO if I remove it, do I remove the config mapping?
	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, customersTenantID)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)
	ingress := k8s.NewIngress(microservice)
	networkPolicy := k8s.NewNetworkPolicy(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)
	// TODO add rawDataLog Configmap
	//configBusinessMoments := businessmomentsadaptor.NewBusinessMomentsConfigmap(microservice)
	ingress.Spec.TLS = k8s.AddIngressTLS([]string{host}, secretName)
	ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.AddIngressRule(host, ingressRules))

	// Could use config-files

	webhookPrefix := strings.ToLower(input.Extra.Ingress.Path)

	container := deployment.Spec.Template.Spec.Containers[0]
	container.ImagePullPolicy = "Always"
	container.Args = []string{
		"raw-data-log",
		"server",
	}
	deployment.Spec.Template.Spec.Containers = []corev1.Container{
		container,
	}

	configEnvVariables.Data = map[string]string{
		"WEBHOOK_REPO":            input.Extra.WriteTo,
		"LISTEN_ON":               "0.0.0.0:8080",
		"WEBHOOK_PREFIX":          webhookPrefix,
		"DOLITTLE_TENANT_ID":      tenant.ID,
		"DOLITTLE_APPLICATION_ID": application.ID,
		"DOLITTLE_ENVIRONMENT":    environment,
		"MICROSERVICE_CONFIG":     "/app/data/microservice_data_from_studio.json",
	}

	if input.Extra.WriteTo == "nats" {
		stanClientID := "ingestor"
		natsServer := fmt.Sprintf("nats.%s.svc.cluster.local", namespace)
		configEnvVariables.Data["NATS_SERVER"] = natsServer
		configEnvVariables.Data["STAN_CLUSTER_ID"] = "stan"
		configEnvVariables.Data["STAN_CLIENT_ID"] = stanClientID
	}

	configFiles.Data = map[string]string{}
	// We store the config data into the config-Files for the service to pick up on
	b, _ := json.MarshalIndent(input, "", "  ")
	configFiles.Data["microservice_data_from_studio.json"] = string(b)

	service.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 8080,
	}
	deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 8080

	// Assuming the namespace exists
	var err error
	client := r.k8sClient
	ctx := context.TODO()

	// ConfigMaps
	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, metav1.CreateOptions{})

	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		// TODO update
		//_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, microserviceConfigmap, metav1.UpdateOptions{})
		fmt.Println("Skipping microserviceConfigmap already exists")
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		fmt.Println("Skipping configEnvVariables already exists")
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		fmt.Println("Skipping configFiles already exists")
	}

	// Secrets
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		fmt.Println("Skipping configSecrets already exists")
	}

	// Ingress
	_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		fmt.Println("Skipping ingress already exists")
	}

	// Service
	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		fmt.Println("Skipping service already exists")
	}

	// NetworkPolicy
	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		fmt.Println("Skipping network policy already exists")
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		fmt.Println("Skipping deployment already exists")
	}

	return nil
}

//func doDolittle(
//	labels map[string]string,
//	annotations map[string]string,
//	action string, namespace string, body []byte, ctx context.Context, cfg *rest.Config) error {
//	obj, dr, err := getKubernetesData(namespace, body, cfg)
//	if err != nil {
//		return err
//	}
//
//	obj.SetNamespace(namespace)
//	currentLabels := obj.GetLabels()
//	if currentLabels == nil {
//		labels = map[string]string{}
//	}
//
//	mergedLabels := k8sLabels.Merge(currentLabels, labels)
//
//	obj.SetLabels(mergedLabels)
//
//	currentAnnotations := obj.GetAnnotations()
//	if currentAnnotations == nil {
//		annotations = map[string]string{}
//	}
//
//	if obj.GetKind() == "Service" {
//		// https://erwinvaneyk.nl/kubernetes-unstructured-to-typed/
//		//var service corev1.Service
//		//err = runtime.DefaultUnstructuredConverter.
//		//	FromUnstructured(unstructured, &cluster)
//		//assertNoError(err)
//
//		//Selector: labels,
//	}
//	fmt.Println(obj.GetKind())
//
//	mergedAnnotations := k8sLabels.Merge(currentAnnotations, annotations)
//	obj.SetAnnotations(mergedAnnotations)
//
//	return doDo(obj, dr, action, ctx)
//}
//
