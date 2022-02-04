package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ErrNotFound      = errors.New("not-found")
	ErrAlreadyExists = errors.New("already-exists")
)

type K8sRepo struct {
	baseConfig *rest.Config
	k8sClient  kubernetes.Interface
	logContext logrus.FieldLogger
	k8sRepoV2  k8s.Repo
	// DO we add logcontext?
}

func NewK8sRepo(k8sClient kubernetes.Interface, config *rest.Config, logContext logrus.FieldLogger) K8sRepo {
	k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))
	return K8sRepo{
		k8sRepoV2:  k8sRepoV2,
		baseConfig: config,
		k8sClient:  k8sClient,
		logContext: logContext,
	}
}

func (r *K8sRepo) GetApplication(applicationID string) (platform.Application, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	ns, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err != nil {
		return platform.Application{}, err
	}

	annotationsMap := ns.GetObjectMeta().GetAnnotations()
	labelMap := ns.GetObjectMeta().GetLabels()

	application := platform.Application{
		Name: labelMap["application"],
		ID:   annotationsMap["dolittle.io/application-id"],
		Tenant: platform.Tenant{
			Name: labelMap["tenant"],
			ID:   annotationsMap["dolittle.io/tenant-id"],
		},
	}

	ingresses, err := r.k8sRepoV2.GetIngresses(namespace)
	if err != nil {
		return platform.Application{}, err
	}

	for _, ingress := range ingresses {
		if len(ingress.Spec.TLS) > 0 {

			labelMap := ingress.GetObjectMeta().GetLabels()
			annotationsMap := ingress.GetObjectMeta().GetAnnotations()

			_, ok := annotationsMap["dolittle.io/application-id"]
			if !ok {
				continue
			}

			for _, rule := range ingress.Spec.Rules {
				for _, rulePath := range rule.IngressRuleValue.HTTP.Paths {
					customerTenantID := ""
					if _, ok := annotationsMap["nginx.ingress.kubernetes.io/configuration-snippet"]; ok {
						customerTenantID = GetCustomerTenantIDFromNginxConfigurationSnippet(annotationsMap["nginx.ingress.kubernetes.io/configuration-snippet"])
					}

					applicationIngress := platform.Ingress{
						Host:             rule.Host,
						Environment:      labelMap["environment"],
						Path:             rulePath.Path,
						CustomerTenantID: customerTenantID,
					}
					application.Ingresses = append(application.Ingresses, applicationIngress)

				}

			}
		}
	}
	return application, nil
}

// GetApplicationNamespaces
// Return a list of namespaces that are "applications"
func (r *K8sRepo) GetApplicationNamespaces() ([]corev1.Namespace, error) {
	client := r.k8sClient
	ctx := context.TODO()
	items, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "tenant,application",
	})
	return items.Items, err
}

// GetApplications Return a list of applications based on customerID
func (r *K8sRepo) GetApplications(customerID string) ([]platform.ShortInfo, error) {
	items, err := r.GetApplicationNamespaces()
	response := make([]platform.ShortInfo, 0)
	if err != nil {
		return response, err
	}

	for _, item := range items {
		annotationsMap := item.GetObjectMeta().GetAnnotations()
		labelMap := item.GetObjectMeta().GetLabels()

		if annotationsMap["dolittle.io/tenant-id"] != customerID {
			continue
		}

		response = append(response, platform.ShortInfo{
			Name: labelMap["application"],
			ID:   annotationsMap["dolittle.io/application-id"],
		})
	}

	return response, nil
}

func (r *K8sRepo) GetMicroservices(applicationID string) ([]platform.MicroserviceInfo, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})

	response := make([]platform.MicroserviceInfo, len(deployments.Items))
	if err != nil {
		return response, err
	}

	for deploymentIndex, deployment := range deployments.Items {
		annotationsMap := deployment.GetObjectMeta().GetAnnotations()
		labelMap := deployment.GetObjectMeta().GetLabels()

		_, ok := labelMap["microservice"]
		if !ok {
			continue
		}

		images := funk.Map(deployment.Spec.Template.Spec.Containers, func(container corev1.Container) platform.ImageInfo {
			return platform.ImageInfo{
				Name:  container.Name,
				Image: container.Image,
			}
		}).([]platform.ImageInfo)

		kind := GetMicroserviceKindFromAnnotations(annotationsMap)

		environment := labelMap["environment"]
		microserviceID := annotationsMap["dolittle.io/microservice-id"]
		ingressURLS, err := r.GetIngressURLsWithCustomerTenantID(applicationID, environment, microserviceID)
		if err != nil {
			return response, err
		}

		ingressHTTPIngressPath, err := r.GetIngressHTTPIngressPath(applicationID, environment, microserviceID)
		if err != nil {
			return response, err
		}

		response[deploymentIndex] = platform.MicroserviceInfo{
			Name:         labelMap["microservice"],
			ID:           microserviceID,
			Environment:  environment,
			Images:       images,
			Kind:         string(kind),
			IngressURLS:  ingressURLS,
			IngressPaths: ingressHTTPIngressPath,
		}
	}

	return response, err
}

func (r *K8sRepo) GetMicroserviceName(applicationID string, environment string, microserviceID string) (string, error) {
	client := r.k8sClient
	ctx := context.TODO()

	namespace := GetApplicationNamespace(applicationID)
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("environment=%s,microservice", environment),
	})

	if err != nil {
		return "", err
	}

	for _, deployment := range deployments.Items {
		if deployment.ObjectMeta.Annotations["dolittle.io/microservice-id"] == microserviceID {
			return deployment.Name, nil
		}
	}
	return "", ErrNotFound
}

func (r *K8sRepo) GetPodStatus(applicationID string, environment string, microserviceID string) (platform.PodData, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	response := platform.PodData{
		Namespace: namespace,
		Microservice: platform.ShortInfo{
			Name: "",
			ID:   microserviceID,
		},
		Pods: []platform.PodInfo{},
	}

	if err != nil {
		return response, err
	}

	// TODO name will be blank if there are no pods, getting name from the json file would help here
	for _, pod := range pods.Items {
		annotationsMap := pod.GetObjectMeta().GetAnnotations()
		labelMap := pod.GetObjectMeta().GetLabels()

		if annotationsMap["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		if strings.ToLower(labelMap["environment"]) != environment {
			continue
		}

		response.Microservice.Name = labelMap["microservice"]
		// Interesting pod.Status.StartTime.String() might not be the same as pod.CreationTimestamp.Time
		age := time.Since(pod.CreationTimestamp.Time)
		started := "N/A"
		if pod.Status.StartTime != nil {
			started = pod.Status.StartTime.String()
		}

		containers := funk.Map(pod.Status.ContainerStatuses, func(container corev1.ContainerStatus) platform.ContainerStatusInfo {
			// Not sure about this logic, I almost want to drop to the cli :P
			// Much work to do here, to figure out the combinations we actually want to support, this will not be good enough
			state := "waiting"

			if *container.Started {
				state = "starting"
			}

			if container.Ready {
				state = "running"
			}

			return platform.ContainerStatusInfo{
				Name:     container.Name,
				Image:    container.Image,
				Restarts: container.RestartCount,
				Started:  started,
				Age:      age.String(),
				State:    state,
			}
		}).([]platform.ContainerStatusInfo)

		response.Pods = append(response.Pods, platform.PodInfo{
			Phase:      string(pod.Status.Phase),
			Name:       pod.Name,
			Containers: containers,
		})
	}

	return response, err
}

// TODO get logs from the pods
func (r *K8sRepo) GetLogs(applicationID string, containerName string, podName string) (string, error) {
	client := r.k8sClient
	ctx := context.TODO()

	namespace := GetApplicationNamespace(applicationID)

	count := int64(100)
	podLogOptions := corev1.PodLogOptions{
		Container: containerName,
		Follow:    false,
		TailLines: &count,
	}

	req := client.CoreV1().Pods(namespace).GetLogs(podName, &podLogOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return err.Error(), nil
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return err.Error(), nil
	}
	str := buf.String()

	return str, nil
}

// CanModifyApplication confirm user is in the tenant and application and if not set the http response
func (r *K8sRepo) CanModifyApplicationWithResponse(w http.ResponseWriter, tenantID string, applicationID string, userID string) bool {
	if tenantID == "" || userID == "" {
		// If the middleware is enabled this shouldn't happen
		utils.RespondWithError(w, http.StatusForbidden, "Tenant-ID and User-ID is missing from the headers")
		return false
	}

	allowed, err := r.CanModifyApplication(tenantID, applicationID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return false
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
		return false
	}

	return true
}

// See https://app.asana.com/0/1200181647276434/1201135519043139 for moving this data to the json files
func (r *K8sRepo) GetMicroserviceDNS(applicationID string, microserviceID string) (string, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	services, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	found := false
	var foundService corev1.Service

	for _, service := range services.Items {
		_, ok := service.ObjectMeta.Labels["microservice"]
		if !ok {
			continue
		}

		if service.ObjectMeta.Annotations["dolittle.io/microservice-id"] == microserviceID {
			found = true
			foundService = service
			break
		}
	}

	if !found {
		return "", fmt.Errorf("no DNS found in applications %s for microservice %s", applicationID, microserviceID)
	}

	// the "svc.cluster.local" postfix might not work for all clusters
	return fmt.Sprintf("%s.application-%s.svc.cluster.local", foundService.Name, applicationID), nil
}

func (r *K8sRepo) GetConfigMap(applicationID string, name string) (*corev1.ConfigMap, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return configMap, err
	}
	return configMap, nil
}

func (r *K8sRepo) GetSecret(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.Secret, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			logContext.WithFields(logrus.Fields{
				"error":      err,
				"secretName": name,
			}).Error("issue talking to cluster")
			return secret, err
		}

		logContext.WithFields(logrus.Fields{
			"error":      err,
			"secretName": name,
		}).Error("secret not found")

		return secret, ErrNotFound
	}
	return secret, nil
}

func (r *K8sRepo) GetServiceAccount(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.ServiceAccount, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	serviceAccount, err := client.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			logContext.WithFields(logrus.Fields{
				"error": err,
				"name":  name,
			}).Error("issue talking to cluster")
			return serviceAccount, err
		}

		logContext.WithFields(logrus.Fields{
			"error": err,
			"name":  name,
		}).Error("service account not found")

		return serviceAccount, ErrNotFound
	}
	return serviceAccount, nil
}

func (r *K8sRepo) CreateServiceAccountFromResource(logContext logrus.FieldLogger, resource *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := resource.Namespace
	newAccount, err := client.CoreV1().ServiceAccounts(namespace).Create(ctx, resource, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("issue talking to cluster")
			return newAccount, err
		}

		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Debug("service account already exists")

		return newAccount, ErrAlreadyExists
	}
	return newAccount, nil
}

func (r *K8sRepo) CreateServiceAccount(logger logrus.FieldLogger, customerID string, customerName string, applicationID string, applicationName string, serviceAccountName string) (*corev1.ServiceAccount, error) {
	namespace := GetApplicationNamespace(applicationID)
	logContext := logger.WithFields(logrus.Fields{
		"customerID":     customerID,
		"applicationID":  applicationID,
		"namespace":      namespace,
		"serviceAccount": serviceAccountName,
		"method":         "CreateServiceAccount",
	})

	serviceAccount := NewServiceAccountResource(serviceAccountName, customerID, customerName, applicationID, applicationName)
	return r.CreateServiceAccountFromResource(logContext, &serviceAccount)
}

func (r *K8sRepo) AddServiceAccountToRoleBinding(logger logrus.FieldLogger, applicationID string, roleBinding string, serviceAccount string) (*rbacv1.RoleBinding, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	logContext := logger.WithFields(logrus.Fields{
		"namespace":   namespace,
		"rolebinding": roleBinding,
	})

	k8sRoleBinding, err := client.RbacV1().RoleBindings(namespace).Get(ctx, roleBinding, metav1.GetOptions{})
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("couldn't find the rolebinding")
		return k8sRoleBinding, err
	}

	for _, subject := range k8sRoleBinding.Subjects {
		// if the serviceaccount already exists in the rolebinding we don't need to update
		if subject.Name == serviceAccount {
			return k8sRoleBinding, ErrAlreadyExists
		}
	}

	k8sRoleBinding.Subjects = append(k8sRoleBinding.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      serviceAccount,
		Namespace: namespace,
	})

	updatedRoleBinding, err := client.RbacV1().RoleBindings(namespace).Update(ctx, k8sRoleBinding, metav1.UpdateOptions{})
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":          err,
			"serviceAccount": serviceAccount,
		}).Error("couldn't update the rolebinding with the serviceaccount")
		return updatedRoleBinding, err
	}

	return updatedRoleBinding, err
}

// TODO this is very specific, lets rename it to be, or make it more generic
// CreateRoleBinding creates a RoleBinding with the given name for the specified Role with empty Subjects
func (r *K8sRepo) CreateRoleBinding(logger logrus.FieldLogger, customerID, customerName, applicationID, applicationName, roleBinding, role string) (*rbacv1.RoleBinding, error) {
	//client := r.k8sClient
	//ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	logContext := logger.WithFields(logrus.Fields{
		"namespace":     namespace,
		"rolebinding":   roleBinding,
		"role":          role,
		"customerID":    customerID,
		"applicationID": applicationID,
		"method":        "CreateRoleBinding",
	})

	resource := NewRoleBindingWithoutSubjects(roleBinding, role, customerID, customerName, applicationID, applicationName)
	return r.CreateRoleBindingFromResource(logContext, &resource)
}

// CanModifyApplication confirm user is in the tenant and application
// Only works when we can use the namespace
func (r *K8sRepo) CanModifyApplication(tenantID string, applicationID string, userID string) (bool, error) {
	attribute := authv1.ResourceAttributes{
		Namespace: fmt.Sprintf("application-%s", applicationID),
		Verb:      "list",
		Resource:  "pods",
	}
	return r.CanModifyApplicationWithResourceAttributes(tenantID, applicationID, userID, attribute)
}

// CanModifyApplicationWithResourceAttributes confirm user is in the tenant and application
// Only works when we can use the namespace
// TODO bringing online the ad group from microsoft will allow us to check group access
func (r *K8sRepo) CanModifyApplicationWithResourceAttributes(tenantID string, applicationID string, userID string, attribute authv1.ResourceAttributes) (bool, error) {
	config := r.GetRestConfig()

	config.Impersonate = rest.ImpersonationConfig{
		UserName: userID,
		Groups: []string{
			platform.GetCustomerGroup(tenantID),
		},
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	selfCheck := authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &attribute,
		},
	}

	resp, err := client.AuthorizationV1().
		SelfSubjectAccessReviews().
		Create(context.TODO(), &selfCheck, metav1.CreateOptions{})

	if err != nil {
		// TODO do we hide this error and log it?
		return false, err
	}

	return resp.Status.Allowed, nil
}

func (r *K8sRepo) GetRestConfig() *rest.Config {
	return rest.CopyConfig(r.baseConfig)
}

func (r *K8sRepo) GetIngressURLsWithCustomerTenantID(applicationID string, environment string, microserviceID string) ([]platform.IngressURLWithCustomerTenantID, error) {
	namespace := GetApplicationNamespace(applicationID)
	urls := make([]platform.IngressURLWithCustomerTenantID, 0)

	ingresses, err := r.k8sRepoV2.GetIngressesByEnvironmentWithMicoservices(namespace, environment)

	if err != nil {
		return urls, err
	}

	for _, ingress := range ingresses {
		annotationsMap := ingress.GetAnnotations()
		if annotationsMap["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		customerTenantID := ""
		if _, ok := annotationsMap["nginx.ingress.kubernetes.io/configuration-snippet"]; ok {
			customerTenantID = GetCustomerTenantIDFromNginxConfigurationSnippet(annotationsMap["nginx.ingress.kubernetes.io/configuration-snippet"])
		}

		for _, rule := range ingress.Spec.Rules {
			for _, rulePath := range rule.HTTP.Paths {
				url := fmt.Sprintf("https://%s%s", rule.Host, rulePath.Path)
				urls = append(urls, platform.IngressURLWithCustomerTenantID{
					URL:              url,
					CustomerTenantID: customerTenantID,
				})
			}
		}
	}

	return urls, nil
}

// GetIngressHTTPIngressPath Return unique Ingress Paths
func (r *K8sRepo) GetIngressHTTPIngressPath(applicationID string, environment string, microserviceID string) ([]networkingv1.HTTPIngressPath, error) {
	namespace := GetApplicationNamespace(applicationID)
	items := make([]networkingv1.HTTPIngressPath, 0)

	ingresses, err := r.k8sRepoV2.GetIngressesByEnvironmentWithMicoservices(namespace, environment)

	if err != nil {
		return items, err
	}

	for _, ingress := range ingresses {
		annotationsMap := ingress.GetAnnotations()
		if annotationsMap["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		for _, rule := range ingress.Spec.Rules {
			for _, rulePath := range rule.HTTP.Paths {
				// For now, there is no need to expose it
				rulePath.Backend = networkingv1.IngressBackend{}

				exists := funk.Contains(items, func(item networkingv1.HTTPIngressPath) bool {
					return rulePath == item
				})

				if exists {
					continue
				}
				items = append(items, rulePath)
			}
		}
	}

	return items, nil
}

func GetApplicationNamespace(id string) string {
	return fmt.Sprintf("application-%s", id)
}

// TODO which is better?
//var tenantFromIngressAnnotationExtractor = regexp.MustCompile(`proxy_set_header\s+Tenant-ID\s+"([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})"`)
//
//func tryGetTenantFromIngress(ingress netv1.Ingress) (bool, platform.TenantId) {
//	tenantHeaderAnnotation := ingress.GetObjectMeta().GetAnnotations()["nginx.ingress.kubernetes.io/configuration-snippet"]
//	tenantID := tenantFromIngressAnnotationExtractor.FindStringSubmatch(tenantHeaderAnnotation)
//	if tenantID == nil {
//		return false, ""
//	}
//	return true, platform.TenantId(tenantID[1])
//}
// This input can have multiple lines
func GetCustomerTenantIDFromNginxConfigurationSnippet(input string) string {
	r, _ := regexp.Compile(`proxy_set_header Tenant-ID "(\S+)"`)
	matches := r.FindStringSubmatch(input)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func (r *K8sRepo) RestartMicroservice(applicationID string, environment string, microserviceID string) error {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	if err != nil {
		return err
	}

	podsToRestart := make([]string, 0)
	for _, pod := range pods.Items {
		annotationsMap := pod.GetObjectMeta().GetAnnotations()
		labelMap := pod.GetObjectMeta().GetLabels()

		if annotationsMap["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		if strings.ToLower(labelMap["environment"]) != environment {
			continue
		}
		podsToRestart = append(podsToRestart, pod.Name)
	}

	for _, podName := range podsToRestart {
		err := client.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *K8sRepo) WriteConfigMap(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := configMap.ObjectMeta.Namespace
	return client.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
}

func (r *K8sRepo) WriteSecret(secret *corev1.Secret) (*corev1.Secret, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := secret.ObjectMeta.Namespace
	return client.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
}

// TODO move once resources land
func GetMicroserviceEnvironmentVariableConfigmapName(name string) string {
	return strings.ToLower(
		fmt.Sprintf("%s-env-variables",
			name,
		),
	)
}

// TODO move once resources land
func GetMicroserviceEnvironmentVariableSecretName(name string) string {
	return strings.ToLower(
		fmt.Sprintf("%s-secret-env-variables",
			name,
		),
	)
}

func GetMicroserviceKindFromAnnotations(annotations map[string]string) (kind platform.MicroserviceKind) {
	kind = platform.MicroserviceKindUnknown
	if kindString, ok := annotations["dolittle.io/microservice-kind"]; ok {
		kind = platform.MicroserviceKind(kindString)
	}
	return
}

func (r *K8sRepo) CreateNamesapce(tenant platform.Tenant, application platform.Application) error {
	client := r.k8sClient
	ctx := context.TODO()
	name := fmt.Sprintf("application-%s", application.ID)

	labels := map[string]string{
		"tenant":      tenant.Name,
		"application": application.Name,
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	return err
}

func (r *K8sRepo) CreateRoleBindingFromResource(logContext logrus.FieldLogger, resource *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := resource.Namespace
	createdRoleBinding, err := client.RbacV1().RoleBindings(namespace).Create(ctx, resource, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("issue talking to cluster")
			return createdRoleBinding, err
		}
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Debug("RoleBinding already exists")
		return createdRoleBinding, ErrAlreadyExists
	}

	return createdRoleBinding, nil
}

func (r *K8sRepo) AddPolicyRule(roleName string, applicationID string, newRule rbacv1.PolicyRule) error {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)

	role, err := client.RbacV1().Roles(namespace).Get(ctx, roleName, metav1.GetOptions{})
	if err != nil {

		return err
	}

	for _, rule := range role.Rules {
		if equality.Semantic.DeepDerivative(rule, newRule) {
			// Nothing to change
			return nil
		}
	}

	role.Rules = append(role.Rules, newRule)

	_, err = client.RbacV1().Roles(namespace).Update(ctx, role, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (r *K8sRepo) RemovePolicyRule(roleName string, applicationID string, newRule rbacv1.PolicyRule) error {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := GetApplicationNamespace(applicationID)

	role, err := client.RbacV1().Roles(namespace).Get(ctx, roleName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	newRules := make([]rbacv1.PolicyRule, 0)
	for _, rule := range role.Rules {
		if !equality.Semantic.DeepDerivative(rule, newRule) {
			newRules = append(newRules, rule)
			continue
		}
	}

	role.Rules = newRules
	_, err = client.RbacV1().Roles(namespace).Update(ctx, role, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
