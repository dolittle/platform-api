package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/google/uuid"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/labels"
)

type MicroserviceResources struct {
	Service                    *corev1.Service
	Deployment                 *appsv1.Deployment
	DolittleConfig             *corev1.ConfigMap
	ConfigFiles                *corev1.ConfigMap
	ConfigEnvironmentVariables *corev1.ConfigMap
	SecretEnvironmentVariables *corev1.Secret
	RbacPolicyRules            []rbacv1.PolicyRule
	IngressResources           *IngressResources
}

type IngressResources struct {
	NetworkPolicy *networkingv1.NetworkPolicy
	Ingresses     []*networkingv1.Ingress
}

type k8sRepo struct {
	k8sClient       kubernetes.Interface
	k8sRepoV2       k8s.Repo
	k8sDolittleRepo platformK8s.K8sRepo
	kind            platform.MicroserviceKind
	isProduction    bool
}

func NewSimpleRepo(k8sClient kubernetes.Interface, k8sDolittleRepo platformK8s.K8sRepo, k8sRepoV2 k8s.Repo, isProduction bool) simple.Repo {
	return k8sRepo{
		k8sClient:       k8sClient,
		k8sRepoV2:       k8sRepoV2,
		k8sDolittleRepo: k8sDolittleRepo,
		kind:            platform.MicroserviceKindSimple,
		isProduction:    isProduction,
	}
}

func (r k8sRepo) Create(namespace string, tenant dolittleK8s.Tenant, application dolittleK8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputSimpleInfo) error {
	var err error

	client := r.k8sClient
	ctx := context.TODO()

	applicationID := application.ID

	resources := NewResources(r.isProduction, namespace, tenant, application, customerTenants, input)

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, resources.DolittleConfig, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, resources.ConfigEnvironmentVariables, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, resources.ConfigFiles, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config files") }) != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(namespace).Create(ctx, resources.SecretEnvironmentVariables, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, resources.Service, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("service") }) != nil {
		return err
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, resources.Deployment, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	// Update developer
	for _, policyRule := range resources.RbacPolicyRules {
		err = r.k8sDolittleRepo.AddPolicyRule("developer", applicationID, policyRule)
		// Not sure this is the best error checking
		if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("policy rule") }) != nil {
			return err
		}
	}

	if input.Extra.Ispublic {
		for _, ingress := range resources.IngressResources.Ingresses {
			_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
			if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("ingress") }) != nil {
				return err
			}
		}

		_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, resources.IngressResources.NetworkPolicy, metav1.CreateOptions{})
		if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("network policy") }) != nil {
			return err
		}
	}

	return nil
}

func (r k8sRepo) Delete(applicationID, environment, microserviceID string) error {
	ctx := context.TODO()
	namespace := platformK8s.GetApplicationNamespace(applicationID)

	deployment, err := r.k8sRepoV2.GetDeployment(namespace, environment, microserviceID)
	if err != nil {
		return err
	}

	// TODO can i get the name? it should be the deployment name
	// Label Environment micoservice
	microserviceName := deployment.Labels["Microservice"]
	policyRules := microserviceK8s.NewMicroservicePolicyRules(microserviceName, environment)

	if err = microserviceK8s.K8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}

	// Selector information for microservice, based on labels
	listOpts := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	// TODO I wonder if the order matters
	if err = microserviceK8s.K8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteIngresses(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteNetworkPolicies(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	// Remove policy rules from developer
	// This might not be the best way if we change things and use this function for cleaning up, but it works
	for _, policyRule := range policyRules {
		err := r.k8sDolittleRepo.RemovePolicyRule("developer", applicationID, policyRule)
		if err != nil {
			return err
		}
	}

	if err = microserviceK8s.K8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}

// Subscribe implements simple.Repo
func (r k8sRepo) Subscribe(customerID, applicationID, environment, microserviceID, tenantID, producerMicroserviceID, producerTenantID, publicStream, partition, scope string) error {
	return r.SubscribeToAnotherApplication(customerID, applicationID, environment, microserviceID, tenantID, producerMicroserviceID, producerTenantID, publicStream, partition, scope, applicationID, environment)
}

// SubscribeToAnotherApplication implements simple.Repo
func (r k8sRepo) SubscribeToAnotherApplication(
	customerID,
	applicationID,
	environment,
	microserviceID,
	tenantID,
	producerMicroserviceID,
	producerTenantID,
	publicStream,
	partition,
	scope,
	producerApplicationID,
	producerEnvironment string) error {
	ctx := context.TODO()
	namespace := platformK8s.GetApplicationNamespace(applicationID)

	// make sure that the producer's namespace is owned by the same customer
	producerNamespaceName := platformK8s.GetApplicationNamespace(producerApplicationID)
	producerNamespace, err := r.k8sClient.CoreV1().Namespaces().Get(ctx, producerNamespaceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	producerCustomerID := producerNamespace.Annotations["dolittle.io/tenant-id"]
	if producerCustomerID != customerID {
		return fmt.Errorf("can't create event horizon subscriptions between different customers, consumer: %s and producer: %s", customerID, producerCustomerID)
	}

	// get the producers microservice
	producerDeployment, err := r.k8sRepoV2.GetDeployment(producerNamespaceName, producerEnvironment, producerMicroserviceID)
	if err != nil {
		return fmt.Errorf("failed to get producers deployment: %w", err)
	}

	consumerDeployment, err := r.k8sRepoV2.GetDeployment(namespace, environment, microserviceID)
	if err != nil {
		return fmt.Errorf("failed to get consumers deployment: %w", err)
	}

	if applicationID != producerApplicationID {
		//make sure the producer has a networkpolicy allowing this microservice to talk to it
		var producerNetworkPolicy *networkingv1.NetworkPolicy
		networkPolicyList, err := r.k8sClient.NetworkingV1().NetworkPolicies(producerNamespaceName).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list producers networkpolicies: %w", err)
		}

		for _, networkPolicy := range networkPolicyList.Items {
			if !strings.EqualFold(networkPolicy.Labels["environment"], producerEnvironment) {
				continue
			}
			if networkPolicy.Annotations["dolittle.io/microservice-id"] != producerMicroserviceID {
				continue
			}
			if !strings.HasSuffix(networkPolicy.Name, "-event-horizons") {
				continue
			}
			producerNetworkPolicy = &networkPolicy
			break
		}

		if producerNetworkPolicy == nil {
			// create it

			networkPolicy := &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf("%s-event-horizons", producerDeployment.Name),
					Namespace:   producerNamespaceName,
					Annotations: producerDeployment.Annotations,
					Labels:      producerDeployment.Labels,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: producerDeployment.Labels,
					},
					PolicyTypes: []networkingv1.PolicyType{"Ingress"},
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							From: []networkingv1.NetworkPolicyPeer{
								{
									NamespaceSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"tenant":      consumerDeployment.Labels["tenant"],
											"application": consumerDeployment.Labels["application"],
										},
									},
									PodSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"environment":  consumerDeployment.Labels["environment"],
											"microservice": consumerDeployment.Labels["microservice"],
										},
									},
								},
							},
						},
					},
				},
			}

			if _, err := r.k8sClient.NetworkingV1().NetworkPolicies(producerNamespaceName).Create(ctx, networkPolicy, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("failed to create network policy for producer %w", err)
			}
		} else {
			// check if the producer already has a networkpolicy for this consumer
			hasIngressRule := false
			for _, ingressRule := range producerNetworkPolicy.Spec.Ingress {
				for _, peer := range ingressRule.From {
					namespaceLabels := peer.NamespaceSelector.MatchLabels
					podLabels := peer.PodSelector.MatchLabels

					if namespaceLabels["tenant"] == consumerDeployment.Labels["tenant"] &&
						namespaceLabels["application"] == consumerDeployment.Labels["application"] &&
						podLabels["environment"] == consumerDeployment.Labels["environment"] &&
						podLabels["microservice"] == consumerDeployment.Labels["microservice"] {
						// found it
						hasIngressRule = true
						break
					}
				}
			}

			if !hasIngressRule {
				ingressRule := networkingv1.NetworkPolicyPeer{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"tenant":      consumerDeployment.Labels["tenant"],
							"application": consumerDeployment.Labels["application"],
						},
					},
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"environment":  consumerDeployment.Labels["environment"],
							"microservice": consumerDeployment.Labels["microservice"],
						},
					},
				}

				producerNetworkPolicy.Spec.Ingress[0].From = append(producerNetworkPolicy.Spec.Ingress[0].From, ingressRule)
				_, err := r.k8sClient.NetworkingV1().NetworkPolicies(producerNamespaceName).Update(ctx, producerNetworkPolicy, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("failed to update the networkpolicy for the producer: %w", err)
				}
			}
		}
	}

	// get producers service
	var producerService *corev1.Service
	serviceList, err := r.k8sClient.CoreV1().Services(producerNamespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list producers services: %w", err)
	}
	for _, service := range serviceList.Items {
		if !strings.EqualFold(service.Labels["environment"], producerEnvironment) {
			continue
		}
		if service.Annotations["dolittle.io/microservice-id"] != producerMicroserviceID {
			continue
		}
		producerService = &service
		break
	}

	if producerService == nil {
		return fmt.Errorf("couldn't find the producers service")
	}

	var producerRuntimePort int32
	for _, port := range producerService.Spec.Ports {
		if port.Name == "runtime" {
			producerRuntimePort = port.Port
		}
	}

	// get the producers -dolittle configmap
	producerConfigmaps, err := r.k8sClient.CoreV1().ConfigMaps(producerNamespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var producerConfigmap *corev1.ConfigMap
	for _, configmap := range producerConfigmaps.Items {
		if !strings.EqualFold(configmap.Labels["environment"], producerEnvironment) {
			continue
		}
		if configmap.Annotations["dolittle.io/microservice-id"] != producerMicroserviceID {
			continue
		}
		if !strings.HasSuffix(configmap.Name, "-dolittle") {
			continue
		}
		producerConfigmap = &configmap
		break
	}

	if producerConfigmap == nil {
		return errors.New("producer didn't have a configmap")
	}

	// get the producers event-horizon-consents.json and update it
	var consents dolittleK8s.MicroserviceEventHorizonConsents
	err = json.Unmarshal([]byte(producerConfigmap.Data["event-horizon-consents.json"]), &consents)
	if err != nil {
		return fmt.Errorf("couldn't deserialize event-horizon-consents.json: %w", err)
	}

	if consents == nil {
		consents = dolittleK8s.MicroserviceEventHorizonConsents{}
	}
	consents[producerTenantID] = append(consents[producerTenantID], dolittleK8s.MicroserviceConsent{
		Microservice: microserviceID,
		Tenant:       tenantID,
		Stream:       publicStream,
		Partition:    partition,
		Consent:      uuid.New().String(),
	})

	b, _ := json.MarshalIndent(consents, "", "  ")
	consentsJSON := string(b)
	producerConfigmap.Data["event-horizon-consents.json"] = consentsJSON

	_, err = r.k8sClient.CoreV1().ConfigMaps(producerNamespaceName).Update(ctx, producerConfigmap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update producers event-horizon-consents.json: %w", err)
	}

	// get the consumers -dolittle configmap
	consumerConfigmaps, err := r.k8sClient.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	var consumerConfigmap *corev1.ConfigMap
	for _, configmap := range consumerConfigmaps.Items {
		if !strings.EqualFold(configmap.Labels["environment"], environment) {
			continue
		}
		if configmap.Annotations["dolittle.io/microservice-id"] != microserviceID {
			continue
		}
		if !strings.HasSuffix(configmap.Name, "-dolittle") {
			continue
		}
		consumerConfigmap = &configmap
		break
	}

	if consumerConfigmap == nil {
		return errors.New("consumer didn't have a configmap")
	}

	// get the consumers microservices.json and update it
	var microservicesConfig dolittleK8s.MicroserviceMicroservices
	err = json.Unmarshal([]byte(consumerConfigmap.Data["microservices.json"]), &microservicesConfig)
	if err != nil {
		return fmt.Errorf("couldn't deserialize microservices.json: %w", err)
	}

	if microservicesConfig == nil {
		microservicesConfig = dolittleK8s.MicroserviceMicroservices{}
	}
	microservicesConfig[producerMicroserviceID] = dolittleK8s.MicroserviceMicroservice{
		Host: fmt.Sprintf("%s-application-%s.svc.cluster.local", producerService.Name, producerApplicationID),
		Port: producerRuntimePort,
	}

	b, _ = json.MarshalIndent(microservicesConfig, "", "  ")
	microservicesConfigJSON := string(b)
	consumerConfigmap.Data["microservices.json"] = microservicesConfigJSON

	// get the consumers event-horizons.json and update it
	var eventHorizons dolittleK8s.MicroserviceEventHorizons
	err = json.Unmarshal([]byte(consumerConfigmap.Data["event-horizons.json"]), &eventHorizons)
	if err != nil {
		return fmt.Errorf("couldn't deserialize event-horizons.json: %w", err)
	}

	if eventHorizons == nil {
		eventHorizons = dolittleK8s.MicroserviceEventHorizons{}
	}

	eventHorizons[tenantID] = append(eventHorizons[tenantID], dolittleK8s.MicroserviceEventHorizon{
		Scope:        scope,
		Microservice: producerMicroserviceID,
		Tenant:       producerTenantID,
		Stream:       publicStream,
		Partition:    partition,
	})

	b, _ = json.MarshalIndent(eventHorizons, "", "  ")
	eventHorizonsJSON := string(b)
	consumerConfigmap.Data["event-horizons.json"] = eventHorizonsJSON

	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Update(ctx, consumerConfigmap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update consumers microservice.json: %w", err)
	}

	return nil
}
