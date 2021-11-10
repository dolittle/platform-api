package platform

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	NotFound = errors.New("not-found")
)

type K8sRepo struct {
	baseConfig *rest.Config
	k8sClient  kubernetes.Interface
}

func NewK8sRepo(k8sClient kubernetes.Interface, config *rest.Config) K8sRepo {

	return K8sRepo{
		baseConfig: config,
		k8sClient:  k8sClient,
	}
}

//annotations:
//    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
//    dolittle.io/application-id: c52e450e-4877-47bf-a584-7874c205e2b9
//  labels:
//    tenant: Flokk
//    application: Shepherd

func (r *K8sRepo) GetIngress(applicationID string) (string, error) {
	ctx := context.TODO()
	opts := metav1.ListOptions{
		LabelSelector: "",
	}

	namespace := fmt.Sprintf("application-%s", applicationID)
	ingresses, _ := r.k8sClient.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	for _, ingress := range ingresses.Items {
		if len(ingress.Spec.Rules) > 0 {
			return ingress.Spec.Rules[0].Host, nil
		}
	}

	return "", errors.New("")
}

func (r *K8sRepo) GetApplication(applicationID string) (Application, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
	ns, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err != nil {
		return Application{}, err
	}

	annotationsMap := ns.GetObjectMeta().GetAnnotations()
	labelMap := ns.GetObjectMeta().GetLabels()

	application := Application{
		Name: labelMap["application"],
		ID:   annotationsMap["dolittle.io/application-id"],
		Tenant: Tenant{
			Name: labelMap["tenant"],
			ID:   annotationsMap["dolittle.io/tenant-id"],
		},
	}

	ingresses, err := r.k8sClient.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return Application{}, err
	}

	for _, ingress := range ingresses.Items {
		// TODO I wonder if we actually want this
		if len(ingress.Spec.TLS) > 0 {

			labelMap := ingress.GetObjectMeta().GetLabels()

			//for _, tls := range ingress.Spec.TLS {
			//	for _, host := range tls.Hosts {
			//
			//		exists := funk.Contains(application.Ingresses, func(item Ingress) bool {
			//			return item.Host == host
			//		})
			//
			//		if exists {
			//			continue
			//		}
			//
			//		applicationIngress := Ingress{
			//			Host:        host,
			//			Environment: labelMap["environment"],
			//		}
			//		application.Ingresses = append(application.Ingresses, applicationIngress)
			//	}
			//}

			for _, rule := range ingress.Spec.Rules {
				// TODO this might crash
				for _, rulePath := range rule.IngressRuleValue.HTTP.Paths {
					// TODO could link microservice to backend via service name
					//fmt.Println(rule.Host, rulePath.Path, *rulePath.PathType)

					applicationIngress := Ingress{
						Host:        rule.Host,
						Environment: labelMap["environment"],
						Path:        rulePath.Path,
					}
					application.Ingresses = append(application.Ingresses, applicationIngress)

				}

			}
		}
	}

	// TODO get the ingress hosts currently in use
	return application, nil
}

func (r *K8sRepo) GetApplications(tenantID string) ([]ShortInfo, error) {
	client := r.k8sClient
	ctx := context.TODO()
	items, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	response := make([]ShortInfo, 0)
	if err != nil {
		return response, err
	}

	for _, item := range items.Items {
		annotationsMap := item.GetObjectMeta().GetAnnotations()
		labelMap := item.GetObjectMeta().GetLabels()

		if annotationsMap["dolittle.io/tenant-id"] != tenantID {
			continue
		}

		response = append(response, ShortInfo{
			Name: labelMap["application"],
			ID:   annotationsMap["dolittle.io/application-id"],
		})
	}

	return response, nil
}

func (r *K8sRepo) GetMicroservices(applicationID string) ([]MicroserviceInfo, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})

	response := make([]MicroserviceInfo, len(deployments.Items))
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

		images := funk.Map(deployment.Spec.Template.Spec.Containers, func(container corev1.Container) ImageInfo {
			return ImageInfo{
				Name:  container.Name,
				Image: container.Image,
			}
		}).([]ImageInfo)

		kind := ""
		if hasKind, ok := annotationsMap["dolittle.io/microservice-kind"]; ok {
			kind = hasKind
		}

		response[deploymentIndex] = MicroserviceInfo{
			Name:        labelMap["microservice"],
			ID:          annotationsMap["dolittle.io/microservice-id"],
			Environment: labelMap["environment"],
			Images:      images,
			Kind:        kind,
		}
	}

	return response, err
}

func (r *K8sRepo) GetMicroserviceName(applicationID string, microserviceID string) (string, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	found := false
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
		return "", errors.New("not-found")
	}

	return foundDeployment.Name, err
}

func (r *K8sRepo) GetPodStatus(applicationID string, microserviceID string, environment string) (PodData, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	response := PodData{
		Namespace: namespace,
		Microservice: ShortInfo{
			Name: "",
			ID:   microserviceID,
		},
		Pods: []PodInfo{},
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

		containers := funk.Map(pod.Status.ContainerStatuses, func(container corev1.ContainerStatus) ContainerStatusInfo {
			// Not sure about this logic, I almost want to drop to the cli :P
			// Much work to do here, to figure out the combinations we actually want to support, this will not be good enough
			state := "waiting"

			if *container.Started {
				state = "starting"
			}

			if container.Ready {
				state = "running"
			}

			return ContainerStatusInfo{
				Name:     container.Name,
				Image:    container.Image,
				Restarts: container.RestartCount,
				Started:  started,
				Age:      age.String(),
				State:    state,
			}
		}).([]ContainerStatusInfo)

		response.Pods = append(response.Pods, PodInfo{
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

	namespace := fmt.Sprintf("application-%s", applicationID)

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
	namespace := fmt.Sprintf("application-%s", applicationID)
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
	namespace := fmt.Sprintf("application-%s", applicationID)
	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return configMap, err
	}
	return configMap, nil
}

func (r *K8sRepo) GetSecret(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.Secret, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
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

		return secret, NotFound
	}
	return secret, nil
}

func (r *K8sRepo) GetServiceAccount(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.ServiceAccount, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
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

		return serviceAccount, NotFound
	}
	return serviceAccount, nil
}

func (r *K8sRepo) CreateServiceAccount(logger logrus.FieldLogger, customerID string, customerName string, applicationID string, applicationName string, serviceAccountName string) (*corev1.ServiceAccount, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
	logContext := logger.WithFields(logrus.Fields{
		"customerID":     customerID,
		"applicationID":  applicationID,
		"namespace":      namespace,
		"serviceAccount": serviceAccountName,
		"method":         "CreateServiceAccount",
	})
	// @joel I could also fetch all of this info from k8s namespace object instead of having to pass it in
	annotations := map[string]string{
		"dolittle.io/tenant-id":      customerID,
		"dolittle.io/applciation-id": applicationID,
	}
	labels := map[string]string{
		"tenant":      customerName,
		"application": applicationName,
	}

	serviceAccount := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceAccountName,
			Namespace:   namespace,
			Annotations: annotations,
			Labels:      labels,
		},
	}

	newAccount, err := client.CoreV1().ServiceAccounts(namespace).Create(ctx, &serviceAccount, metav1.CreateOptions{})
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

		return newAccount, nil
	}
	return newAccount, nil
}

func (r *K8sRepo) AddServiceAccountToRoleBinding(logger logrus.FieldLogger, applicationID string, roleBinding string, serviceaccount string) (*rbacv1.RoleBinding, error) {
	client := r.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
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
		if subject.Name == serviceaccount {
			return k8sRoleBinding, nil
		}
	}

	k8sRoleBinding.Subjects = append(k8sRoleBinding.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      serviceaccount,
		Namespace: namespace,
	})

	updatedRoleBinding, err := client.RbacV1().RoleBindings(namespace).Update(ctx, k8sRoleBinding, metav1.UpdateOptions{})
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":          err,
			"serviceaccount": serviceaccount,
		}).Error("couldn't update the rolebinding with the serviceaccount")
		return updatedRoleBinding, err
	}

	return updatedRoleBinding, err
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
			fmt.Sprintf("tenant-%s", tenantID),
		},
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	selfCheck := authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &attribute,
		},
	}

	resp, err := clientset.AuthorizationV1().
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
