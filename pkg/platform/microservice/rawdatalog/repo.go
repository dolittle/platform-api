package rawdatalog

import (
	"errors"
	"log"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"

	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"k8s.io/client-go/kubernetes"

	"context"
	"encoding/json"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type RawDataLogIngestorRepo struct {
	k8sClient       kubernetes.Interface
	k8sDolittleRepo platform.K8sRepo
	gitRepo         storage.Repo
}

func NewRawDataLogIngestorRepo(k8sDolittleRepo platform.K8sRepo, k8sClient kubernetes.Interface, gitRepo storage.Repo) RawDataLogIngestorRepo {
	return RawDataLogIngestorRepo{
		k8sClient:       k8sClient,
		k8sDolittleRepo: k8sDolittleRepo,
		gitRepo:         gitRepo,
	}
}

func (r RawDataLogIngestorRepo) Exists(namespace string, environment string, microserviceID string) (bool, error) {
	return false, errors.New("TODO")
}

func (r RawDataLogIngestorRepo) Update(namespace string, tenant k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputRawDataLogIngestorInfo) error {
	return errors.New("TODO")
}

func (r RawDataLogIngestorRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputRawDataLogIngestorInfo) error {

	microservice := k8s.Microservice{
		Kind:        platform.MicroserviceKindRawDataLogIngestor,
		ID:          input.Dolittle.MicroserviceID, // TODO: I think the RawDataLogWebhookIngestor should have a fixed ID - not sure if we want to do that here or in the frontend?
		Name:        "raw-data-log-ingestor",
		Environment: input.Environment,
		Application: application,
		Tenant:      tenant,
	}

	labels := k8s.GetLabels(microservice)
	annotations := k8s.GetAnnotations(microservice)

	// TODO changing writeTo will break this.
	if input.Extra.WriteTo != "stdout" {
		action := "upsert"
		if err := r.doNats(namespace, labels, annotations, input, action); err != nil {
			fmt.Println("Could not doNats", err)
			return err
		}
	}

	// TODO add microservice
	err := r.doDolittle(namespace, tenant, application, applicationIngress, input)
	if err != nil {
		fmt.Println("Could not doDolittle", err)
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
	var foundDeployment appsv1.Deployment
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
		LabelSelector: k8slabels.FormatLabels(foundDeployment.GetObjectMeta().GetLabels()),
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
}

func (r RawDataLogIngestorRepo) doStatefulService(namespace string, configMap *corev1.ConfigMap, service *corev1.Service, statfulset *appsv1.StatefulSet, action string) error {
	ctx := context.TODO()

	if action == "delete" {
		if err := r.k8sClient.AppsV1().StatefulSets(namespace).Delete(ctx, statfulset.GetName(), metav1.DeleteOptions{}); err != nil {
			return err
		}
		if err := r.k8sClient.CoreV1().Services(namespace).Delete(ctx, service.GetName(), metav1.DeleteOptions{}); err != nil {
			return err
		}
		if err := r.k8sClient.CoreV1().ConfigMaps(namespace).Delete(ctx, configMap.GetName(), metav1.DeleteOptions{}); err != nil {
			return err
		}
		return nil
	}

	if action != "upsert" {
		return errors.New("action not supported")
	}

	if existing, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMap.GetName(), metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if _, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		configMap.ResourceVersion = existing.ResourceVersion
		if _, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	if existing, err := r.k8sClient.CoreV1().Services(namespace).Get(ctx, service.GetName(), metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if _, err := r.k8sClient.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		service.ResourceVersion = existing.ResourceVersion
		if _, err := r.k8sClient.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	if existing, err := r.k8sClient.AppsV1().StatefulSets(namespace).Get(ctx, statfulset.GetName(), metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if _, err := r.k8sClient.AppsV1().StatefulSets(namespace).Create(ctx, statfulset, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		statfulset.ResourceVersion = existing.ResourceVersion
		if _, err := r.k8sClient.AppsV1().StatefulSets(namespace).Update(ctx, statfulset, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (r RawDataLogIngestorRepo) doNats(namespace string, labels, annotations k8slabels.Set, input platform.HttpInputRawDataLogIngestorInfo, action string) error {

	environment := strings.ToLower(input.Environment)

	natsLabels := k8slabels.Merge(labels, k8slabels.Set{"infrastructure": "Nats"})
	natsLabels["microservice"] = ""
	stanLabels := k8slabels.Merge(labels, k8slabels.Set{"infrastructure": "Stan"})
	stanLabels["microservice"] = ""

	nats := createNatsResources(namespace, environment, natsLabels, annotations)
	stan := createStanResources(namespace, environment, stanLabels, annotations)

	if err := r.doStatefulService(namespace, nats.configMap, nats.service, nats.statfulset, action); err != nil {
		return err
	}
	if err := r.doStatefulService(namespace, stan.configMap, stan.service, stan.statfulset, action); err != nil {
		return err
	}

	return nil
}

func (r RawDataLogIngestorRepo) doDolittle(namespace string, customer k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputRawDataLogIngestorInfo) error {

	// TODO not sure where this comes from really, assume dynamic
	// tenantID := "17426336-fb8e-4425-8ab7-07d488367be9"
	storedApplication, err := r.gitRepo.GetApplication(customer.ID, application.ID)
	if err != nil {
		return err
	}
	tenantID, err := storedApplication.GetTenantForEnvironment(input.Environment)
	if err != nil {
		return err
	}

	environment := strings.ToLower(input.Environment)
	host := applicationIngress.Host
	secretName := applicationIngress.SecretName

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customer,
		Application: application,
		Environment: environment,
		ResourceID:  string(tenantID),
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

	// TODO update should not allow changes to:
	// - name
	// What else?

	// TODO do I need this?
	// TODO if I remove it, do I remove the config mapping?
	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, string(tenantID))
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
		"DOLITTLE_TENANT_ID":      customer.ID,
		"DOLITTLE_APPLICATION_ID": application.ID,
		"DOLITTLE_ENVIRONMENT":    environment,
		"MICROSERVICE_CONFIG":     "/app/data/microservice_data_from_studio.json",
	}

	if input.Extra.WriteTo == "nats" {
		stanClientID := "ingestor"
		// TODO we hardcode nats
		natsServer := fmt.Sprintf("%s-nats.%s.svc.cluster.local", environment, namespace)
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
	client := r.k8sClient
	ctx := context.TODO()

	// ConfigMaps
	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, metav1.CreateOptions{})

	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, microserviceConfigmap, metav1.UpdateOptions{})
		fmt.Println("microserviceConfigmap already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, configEnvVariables, metav1.UpdateOptions{})
		fmt.Println("configEnvVariables already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, configFiles, metav1.UpdateOptions{})
		fmt.Println("configFiles already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// Secrets
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().Secrets(namespace).Update(ctx, configSecrets, metav1.UpdateOptions{})
		fmt.Println("configSecrets already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// Ingress
	_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
		fmt.Println("Ingress already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// Service
	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}
		// TODO this breaks
		// I think I need to be strict about what is changeable

		fmt.Println("Skipping service as it already exists")
		// TODO pretty sure I don't want to update this
		data, err := json.Marshal(service)
		if err != nil {
			return err
		}

		_, err = client.CoreV1().Services(namespace).Patch(ctx, service.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "platform-api",
		})

		//_, err = client.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
		fmt.Println("Service already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// NetworkPolicy
	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		fmt.Println("Network Policy already exists")
		_, err = client.NetworkingV1().NetworkPolicies(namespace).Update(ctx, networkPolicy, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		fmt.Println("Deployment Policy already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	return nil
}
