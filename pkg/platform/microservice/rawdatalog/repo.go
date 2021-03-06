package rawdatalog

import (
	"errors"
	"log"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"

	"github.com/dolittle/platform-api/pkg/platform/customertenant"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"context"
	"encoding/json"
	"fmt"
	"strings"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type RawDataLogIngestorRepo struct {
	k8sClient       kubernetes.Interface
	k8sDolittleRepo platformK8s.K8sRepo
	logContext      logrus.FieldLogger
	isProduction    bool
}

func NewRawDataLogIngestorRepo(isProduction bool, k8sDolittleRepo platformK8s.K8sRepo, k8sClient kubernetes.Interface, logContext logrus.FieldLogger) RawDataLogIngestorRepo {
	return RawDataLogIngestorRepo{
		k8sClient:       k8sClient,
		k8sDolittleRepo: k8sDolittleRepo,
		isProduction:    isProduction,
		logContext:      logContext,
	}
}

// Checks whether
func (r RawDataLogIngestorRepo) Exists(namespace string, environment string) (exists bool, microserviceID string, err error) {
	r.logContext.WithFields(logrus.Fields{
		"namespace":   namespace,
		"environment": environment,
		"method":      "RawDataLogIngestorRepo.Exists",
	}).Debug("Checking for RawDataLog microservices existence")
	ctx := context.TODO()
	deployments, err := r.k8sClient.AppsV1().Deployments(namespace).List(ctx, metaV1.ListOptions{})

	if err != nil {
		return false, "", err
	}

	for _, deployment := range deployments.Items {
		annotations := deployment.GetAnnotations()

		// the microserviceID is unique per microservice so that's enough for the check
		if annotations["dolittle.io/microservice-kind"] == string(platform.MicroserviceKindRawDataLogIngestor) {
			r.logContext.WithFields(logrus.Fields{
				"namespace":   namespace,
				"environment": environment,
				"method":      "RawDataLogIngestorRepo.Exists",
			}).Debug("Found a RawDataLog microservice")
			return true, annotations["dolittle.io/microservice-id"], nil
		}
	}
	r.logContext.WithFields(logrus.Fields{
		"namespace":   namespace,
		"environment": environment,
		"method":      "RawDataLogIngestorRepo.Exists",
	}).Debug("Didn't find a RawDataLog microservice")

	return false, "", nil
}

// Update updates the config files config map of the raw data log microservice
func (r RawDataLogIngestorRepo) Update(namespace string, customer k8s.Tenant, application k8s.Application, input platform.HttpInputRawDataLogIngestorInfo) error {
	logger := r.logContext.WithFields(logrus.Fields{
		"namespace":   namespace,
		"customer":    customer.ID,
		"application": application.ID,
		"environment": input.Environment,
		"method":      "RawDataLogIngestorRepo.Update",
	})
	logger.Debug("Updating the RawDataLog microservice")

	exists, _, err := r.Exists(namespace, input.Environment)
	if err != nil {
		logger.WithError(err).Error("Failed to check if Raw Data Log exists")
		return err
	}
	if !exists {
		logger.Warnf("A Raw Data Log doesn't exist for namespace %s and environment %s", namespace, input.Environment)
		return fmt.Errorf("a Raw Data Log doesn't exist for namespace %s and environment %s", namespace, input.Environment)
	}

	environment := input.Environment

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	kind := input.Kind

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customer,
		Application: application,
		Environment: environment,
		Kind:        kind,
	}

	configFiles := k8s.NewConfigFilesConfigmap(microservice)

	ctx := context.TODO()
	configFiles, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, configFiles.Name, metaV1.GetOptions{})
	if err != nil {
		logger.WithError(err).Error("Could not get config files")
		return err
	}
	configFiles = r.configureConfigFiles(configFiles, input)

	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Update(ctx, configFiles, metaV1.UpdateOptions{})
	if err != nil {
		logger.WithError(err).Error("Could not update config files")
		return err
	}

	return nil
}

func (r RawDataLogIngestorRepo) Create(namespace string, customer k8s.Tenant, application k8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputRawDataLogIngestorInfo) error {
	logger := r.logContext.WithFields(logrus.Fields{
		"namespace":   namespace,
		"customer":    customer.ID,
		"application": application.ID,
		"environment": input.Environment,
		"method":      "RawDataLogIngestorRepo.Create",
	})
	logger.Debug("Starting to create the RawDataLog microservice")

	microservice := k8s.Microservice{
		Kind:        platform.MicroserviceKindRawDataLogIngestor,
		ID:          input.Dolittle.MicroserviceID, // TODO: I think the RawDataLogWebhookIngestor should have a fixed ID - not sure if we want to do that here or in the frontend?
		Name:        "raw-data-log-ingestor",
		Environment: input.Environment,
		Application: application,
		Tenant:      customer,
	}

	labels := k8s.GetLabels(microservice)
	annotations := k8s.GetAnnotations(microservice)

	if len(customerTenants) == 0 {
		return errors.New("no-customer-tenants")
	}

	// TODO changing writeTo will break this.
	if input.Extra.WriteTo != "stdout" {

		action := "upsert"
		if err := r.doNats(namespace, labels, annotations, input, action); err != nil {
			logger.WithError(err).Error("Could not doNats")
			return err
		}
	}

	if err := r.doDolittle(namespace, customer, application, customerTenants, input); err != nil {
		logger.WithError(err).Error("Could not doDolittle")
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
	opts := metaV1.ListOptions{}
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

		// TODO microservice-id isn't unique, it can be shared by microservices accoss environments sadly
		// so this should also check for the correct environment to be 100% sure
		// it probably doesn't matter too much for RawDataLogIngerstorRepo currently
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
		GetScale(ctx, foundDeployment.Name, metaV1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("issue")
	}

	sc := *s
	if sc.Spec.Replicas != 0 {
		sc.Spec.Replicas = 0
		_, err := client.AppsV1().
			Deployments(namespace).
			UpdateScale(ctx, foundDeployment.Name, &sc, metaV1.UpdateOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Selector information for microservice, based on labels
	opts = metaV1.ListOptions{
		LabelSelector: k8slabels.FormatLabels(foundDeployment.GetObjectMeta().GetLabels()),
	}

	// Remove configmaps
	configs, _ := client.CoreV1().ConfigMaps(namespace).List(ctx, opts)

	for _, config := range configs.Items {
		err = client.CoreV1().ConfigMaps(namespace).Delete(ctx, config.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Remove secrets
	secrets, _ := client.CoreV1().Secrets(namespace).List(ctx, opts)
	for _, secret := range secrets.Items {
		err = client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Remove Ingress
	ingresses, _ := client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	for _, ingress := range ingresses.Items {
		err = client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove Network Policy
	policies, _ := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, opts)
	for _, policy := range policies.Items {
		err = client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, policy.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove Service
	services, _ := client.CoreV1().Services(namespace).List(ctx, opts)
	for _, service := range services.Items {
		err = client.CoreV1().Services(namespace).Delete(ctx, service.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove statefulset
	statefulSets, _ := client.AppsV1().StatefulSets(namespace).List(ctx, opts)
	for _, stateful := range statefulSets.Items {
		err = client.AppsV1().StatefulSets(namespace).Delete(ctx, stateful.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove deployment
	err = client.AppsV1().
		Deployments(namespace).
		Delete(ctx, foundDeployment.Name, metaV1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("todo")
	}

	return nil
}

// Creates or deletes the statefulset, service and configmap of the given statefulset, service and configmap
func (r RawDataLogIngestorRepo) doStatefulService(namespace string, configMap *corev1.ConfigMap, service *corev1.Service, statfulset *appsv1.StatefulSet, action string) error {
	ctx := context.TODO()

	if action == "delete" {
		if err := r.k8sClient.AppsV1().StatefulSets(namespace).Delete(ctx, statfulset.GetName(), metaV1.DeleteOptions{}); err != nil {
			return err
		}
		if err := r.k8sClient.CoreV1().Services(namespace).Delete(ctx, service.GetName(), metaV1.DeleteOptions{}); err != nil {
			return err
		}
		if err := r.k8sClient.CoreV1().ConfigMaps(namespace).Delete(ctx, configMap.GetName(), metaV1.DeleteOptions{}); err != nil {
			return err
		}
		return nil
	}

	if action != "upsert" {
		return errors.New("action not supported")
	}

	if existing, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMap.GetName(), metaV1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if _, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metaV1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		configMap.ResourceVersion = existing.ResourceVersion
		if _, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metaV1.UpdateOptions{}); err != nil {
			return err
		}
	}

	if existing, err := r.k8sClient.CoreV1().Services(namespace).Get(ctx, service.GetName(), metaV1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if _, err := r.k8sClient.CoreV1().Services(namespace).Create(ctx, service, metaV1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		service.ResourceVersion = existing.ResourceVersion
		if _, err := r.k8sClient.CoreV1().Services(namespace).Update(ctx, service, metaV1.UpdateOptions{}); err != nil {
			return err
		}
	}

	if existing, err := r.k8sClient.AppsV1().StatefulSets(namespace).Get(ctx, statfulset.GetName(), metaV1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if _, err := r.k8sClient.AppsV1().StatefulSets(namespace).Create(ctx, statfulset, metaV1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// TODO this probably won't work, as it's mostly forbidden to k8s to update the statefulset spec
		statfulset.ResourceVersion = existing.ResourceVersion
		if _, err := r.k8sClient.AppsV1().StatefulSets(namespace).Update(ctx, statfulset, metaV1.UpdateOptions{}); err != nil {
			return err
		}
	}
	r.logContext.WithFields(logrus.Fields{
		"namespace": namespace,
		"method":    "RawDataLogIngestorRepo.doStatefulService",
	}).Debug("Finished creating statefulservice")

	return nil
}

func (r RawDataLogIngestorRepo) doNats(namespace string, labels, annotations k8slabels.Set, input platform.HttpInputRawDataLogIngestorInfo, action string) error {
	r.logContext.WithFields(logrus.Fields{
		"namespace": namespace,
		"method":    "RawDataLogIngestorRepo.doNats",
	}).Debug("Starting to create the nats and stan")

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

// Creates the RawDataLog microservice in k8s
// TODO this tenant is wrong
func (r RawDataLogIngestorRepo) doDolittle(namespace string, customer k8s.Tenant, application k8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputRawDataLogIngestorInfo) error {
	isProduction := r.isProduction
	r.logContext.WithFields(logrus.Fields{
		"namespace": namespace,
		"method":    "RawDataLogIngestorRepo.doDolittle",
	}).Debug("Starting to create RawDataLog microservice")

	environment := input.Environment

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage
	kind := input.Kind

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customer,
		Application: application,
		Environment: environment,
		Kind:        kind,
	}

	// TODO I don't understand why we are creating microserviceconfigmap, yet.
	// TODO update should not allow changes to:
	// - name
	// What else?

	// TODO do I need this?
	// TODO if I remove it, do I remove the config mapping?
	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, customerTenants)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)

	networkPolicy := k8s.NewNetworkPolicy(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)
	// TODO add rawDataLog Configmap
	//configBusinessMoments := businessmomentsadaptor.NewBusinessMomentsConfigmap(microservice)

	// TODO this needs coming back to when / if we want to bring rawdatalog back online
	ingresses := customertenant.CreateIngresses(isProduction, customerTenants, microservice, service.Name, input.Extra.Ingress)
	if len(ingresses) == 0 {
		return errors.New("no ingresses were found")
	}
	ingress := ingresses[0]
	// Could use config-files

	webhookPrefix := strings.ToLower(input.Extra.Ingress.Path)
	deployment = r.configureDeployment(deployment)

	configEnvVariables.Data = map[string]string{
		"WEBHOOK_REPO":            input.Extra.WriteTo,
		"LISTEN_ON":               "0.0.0.0:8080",
		"WEBHOOK_PREFIX":          webhookPrefix,
		"DOLITTLE_TENANT_ID":      customer.ID,
		"DOLITTLE_APPLICATION_ID": application.ID,
		"DOLITTLE_ENVIRONMENT":    strings.ToLower(environment),
		"MICROSERVICE_CONFIG":     "/app/data/microservice_data_from_studio.json",
		"TOPIC":                   "purchaseorders",
	}

	if input.Extra.WriteTo == "nats" {
		stanClientID := "ingestor"
		// TODO we hardcode nats
		natsServer := strings.ToLower(fmt.Sprintf("%s-nats.%s.svc.cluster.local", environment, namespace))
		configEnvVariables.Data["NATS_SERVER"] = natsServer
		configEnvVariables.Data["STAN_CLUSTER_ID"] = "stan"
		configEnvVariables.Data["STAN_CLIENT_ID"] = stanClientID
	}

	configFiles = r.configureConfigFiles(configFiles, input)

	service.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 8080,
	}

	// Assuming the namespace exists
	client := r.k8sClient
	ctx := context.TODO()

	// ConfigMaps
	_, err := client.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, metaV1.CreateOptions{})

	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, microserviceConfigmap, metaV1.UpdateOptions{})
		fmt.Println("microserviceConfigmap already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
			// TODO continuing after an error, not ideal
		}
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, metaV1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, configEnvVariables, metaV1.UpdateOptions{})
		fmt.Println("configEnvVariables already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, metaV1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, configFiles, metaV1.UpdateOptions{})
		fmt.Println("configFiles already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// Secrets
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, metaV1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.CoreV1().Secrets(namespace).Update(ctx, configSecrets, metaV1.UpdateOptions{})
		fmt.Println("configSecrets already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// Ingress
	_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metaV1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metaV1.UpdateOptions{})
		fmt.Println("Ingress already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// Service
	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metaV1.CreateOptions{})
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

		_, err = client.CoreV1().Services(namespace).Patch(ctx, service.GetName(), types.ApplyPatchType, data, metaV1.PatchOptions{
			FieldManager: "platform-api",
		})

		//_, err = client.CoreV1().Services(namespace).Update(ctx, service, metaV1.UpdateOptions{})
		fmt.Println("Service already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	// NetworkPolicy
	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metaV1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		fmt.Println("Network Policy already exists")
		_, err = client.NetworkingV1().NetworkPolicies(namespace).Update(ctx, networkPolicy, metaV1.UpdateOptions{})
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metaV1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			log.Fatal(err)
			return errors.New("issue")
		}

		_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metaV1.UpdateOptions{})
		fmt.Println("Deployment Policy already exists")
		if err != nil {
			fmt.Println("error updating")
			fmt.Println(err.Error())
		}
	}

	r.logContext.WithFields(logrus.Fields{
		"namespace": namespace,
		"method":    "RawDataLogIngestorRepo.doDolittle",
	}).Debug("Finished creating RawDataLog microservice")

	return nil
}

func (r RawDataLogIngestorRepo) configureConfigFiles(configFiles *corev1.ConfigMap, input platform.HttpInputRawDataLogIngestorInfo) *corev1.ConfigMap {
	configFiles.Data = map[string]string{}
	// We store the config data into the config-Files for the service to pick up on
	b, _ := json.MarshalIndent(input, "", "  ")
	configFiles.Data["microservice_data_from_studio.json"] = string(b)
	return configFiles
}

func (r RawDataLogIngestorRepo) configureDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	container := deployment.Spec.Template.Spec.Containers[0]
	container.ImagePullPolicy = "Always"
	container.Args = []string{
		"raw-data-log",
		"server",
	}
	deployment.Spec.Template.Spec.Containers = []corev1.Container{
		container,
	}
	deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 8080
	return deployment
}
