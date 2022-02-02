package k8s

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"

	appsv1 "k8s.io/api/apps/v1"

	v1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/docker/docker/api/types"
)

var (
	AllowNetworkPolicyForMonitoring = networkingv1.NetworkPolicyPeer{
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"system":     "Monitoring",
				"monitoring": "Metrics",
			},
		},
		PodSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"system":     "Monitoring",
				"monitoring": "Metrics",
				"service":    "Prometheus",
			},
		},
	}
	AllowNetworkPolicyForSystemAPI = networkingv1.NetworkPolicyPeer{
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"system": "Api",
			},
		},
		PodSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"system":      "Api",
				"environment": "Prod",
				"service":     "Api-V1",
			},
		},
	}
)

type DockerConfigJSON struct {
	Auths map[string]types.AuthConfig `json:"auths"`
}
type Resources struct {
	Namespace                      *corev1.Namespace
	Acr                            *corev1.Secret
	Storage                        *corev1.Secret
	Environments                   []EnvironmentResources
	Rbac                           RbacResources
	LocalDevRoleBindingToDeveloper *rbacv1.RoleBinding
	ServiceAccounts                []ServiceAccount
}

type ServiceAccount struct {
	Name            string
	RoleBindingName string
	Customer        dolittleK8s.Tenant
	Application     dolittleK8s.Application
}

type MongoSettings struct {
	ShareName       string
	VolumeSize      string
	CronJobSchedule string
}
type MongoResources struct {
	Service     *corev1.Service
	StatefulSet *appsv1.StatefulSet
	Cronjob     *v1beta1.CronJob
}
type RbacResources struct {
	RoleBinding    *rbacv1.RoleBinding
	Role           *rbacv1.Role
	ServiceAccount *corev1.ServiceAccount // I dont think we get this yet, @joel code is where to go
}

type EnvironmentResources struct {
	Mongo               MongoResources
	NetworkPolicy       *networkingv1.NetworkPolicy
	Tenants             *corev1.ConfigMap
	RbacRolePolicyRules []rbacv1.PolicyRule
}

func NewNamespace(tenant dolittleK8s.Tenant, application dolittleK8s.Application) *corev1.Namespace {
	name := fmt.Sprintf("application-%s", application.ID)

	labels := map[string]string{
		"tenant":      platformK8s.ParseLabel(tenant.Name),
		"application": platformK8s.ParseLabel(application.Name),
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
		},
	}
	return namespace
}

func NewServiceAccountsInfo(customer dolittleK8s.Tenant, application dolittleK8s.Application) []ServiceAccount {
	return []ServiceAccount{
		{
			Name:            "devops",
			RoleBindingName: "devops",
			Customer:        customer,
			Application:     application,
		},
	}
}

func NewMongoPortForwardPolicyRole(environment string) rbacv1.PolicyRule {
	cleaned := strings.ToLower(environment)
	portForward := fmt.Sprintf("%s-mongo-0", cleaned)

	return rbacv1.PolicyRule{
		Verbs:         []string{"create"},
		APIGroups:     []string{""},
		Resources:     []string{"pods/portforward"},
		ResourceNames: []string{portForward},
	}
}

func NewDeveloperRole(tenant dolittleK8s.Tenant, application dolittleK8s.Application, azureGroupId string) RbacResources {
	tenantGroup := platform.GetCustomerGroup(tenant.ID)
	namespace := fmt.Sprintf("application-%s", application.ID)
	labels := map[string]string{
		"tenant":      platformK8s.ParseLabel(tenant.Name),
		"application": platformK8s.ParseLabel(application.Name),
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	resource := RbacResources{}
	resource.Role = &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/rbacv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "developer",
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
				APIGroups: []string{""},
				Resources: []string{
					"pods",
					"pods/log",
				},
			},
			{
				Verbs: []string{
					"get",
					"list",
					"watch",
					"patch",
					"update",
				},
				APIGroups: []string{"apps"},
				Resources: []string{
					"deployments",
					"deployments/scale",
				},
			},
			{
				Verbs: []string{
					"get",
					"list",
				},
				APIGroups: []string{""},
				Resources: []string{
					"services",
					"pods",
				},
			},
		},
	}

	resource.RoleBinding = &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/rbacv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "developer",
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "Group",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     azureGroupId,
			},
			{
				Kind:     "Group",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     tenantGroup,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "developer",
		},
	}

	return resource
}

func MakeCustomerAcrDockerConfig(customer platform.TerraformCustomer) string {
	registry := fmt.Sprintf("%s.azurecr.io", customer.ContainerRegistryName)
	auth := fmt.Sprintf("%s:%s", customer.ContainerRegistryUsername, customer.ContainerRegistryPassword)
	authConfigs := map[string]types.AuthConfig{}
	authConfigs[registry] = types.AuthConfig{
		Username: customer.ContainerRegistryUsername,
		Password: customer.ContainerRegistryPassword,
		Auth:     base64.StdEncoding.EncodeToString([]byte(auth)),
	}

	auths := DockerConfigJSON{
		Auths: map[string]types.AuthConfig{
			registry: {
				Username: customer.ContainerRegistryUsername,
				Password: customer.ContainerRegistryPassword,
				Auth:     base64.StdEncoding.EncodeToString([]byte(auth)),
			},
		},
	}
	b, _ := json.Marshal(auths)
	return string(b)
}

func NewAcr(tenant dolittleK8s.Tenant, application dolittleK8s.Application, secretData string) *corev1.Secret {
	namespace := fmt.Sprintf("application-%s", application.ID)
	labels := map[string]string{
		"tenant":      platformK8s.ParseLabel(tenant.Name),
		"application": platformK8s.ParseLabel(application.Name),
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	name := "acr"
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   namespace,
		},
		Type:       corev1.SecretTypeDockerConfigJson,
		StringData: map[string]string{".dockerconfigjson": secretData},
	}
}

func NewStorage(tenant dolittleK8s.Tenant, application dolittleK8s.Application, azureStorageAccountName string, azureStorageAccountKey string) *corev1.Secret {
	namespace := fmt.Sprintf("application-%s", application.ID)
	labels := map[string]string{
		"tenant":      platformK8s.ParseLabel(tenant.Name),
		"application": platformK8s.ParseLabel(application.Name),
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}
	name := "storage-account-secret"
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "corev1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   namespace,
		},
		StringData: map[string]string{
			"azurestorageaccountname": azureStorageAccountName,
			"azurestorageaccountkey":  azureStorageAccountKey,
		},
	}
}

func NewTenantsConfigMap(environment string, tenant dolittleK8s.Tenant, application dolittleK8s.Application, customerTenants []platform.CustomerTenantInfo) *corev1.ConfigMap {
	namespace := fmt.Sprintf("application-%s", application.ID)
	name := fmt.Sprintf("%s-tenants", strings.ToLower(environment))

	labels := map[string]string{
		"tenant":      platformK8s.ParseLabel(tenant.Name),
		"application": platformK8s.ParseLabel(application.Name),
		"environment": platformK8s.ParseLabel(environment),
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	tenants := make(platform.RuntimeTenantsIDS)

	for _, customerTenant := range customerTenants {
		// Ugly stuff, but matches today
		empty := make(map[string]interface{})
		tenants[customerTenant.CustomerTenantID] = empty
	}

	b, _ := json.MarshalIndent(tenants, "", "  ")
	tenantsJSON := string(b)

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "corev1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Data: map[string]string{"tenants.json": tenantsJSON},
	}
}

func NewEnvironment(environment string, tenant dolittleK8s.Tenant, application dolittleK8s.Application, mongoSettings MongoSettings, customerTenants []platform.CustomerTenantInfo) EnvironmentResources {
	resources := EnvironmentResources{
		Mongo:               NewMongo(environment, tenant, application, mongoSettings),
		NetworkPolicy:       NewNetworkPolicy(environment, tenant, application),
		Tenants:             NewTenantsConfigMap(environment, tenant, application, customerTenants),
		RbacRolePolicyRules: make([]rbacv1.PolicyRule, 0),
	}

	// Add port-forward rule
	resources.RbacRolePolicyRules = append(resources.RbacRolePolicyRules, NewMongoPortForwardPolicyRole(environment))

	return resources
}

func NewNetworkPolicy(environment string, tenant dolittleK8s.Tenant, application dolittleK8s.Application) *networkingv1.NetworkPolicy {
	namespace := fmt.Sprintf("application-%s", application.ID)
	name := strings.ToLower(environment)

	labels := map[string]string{
		"tenant":      platformK8s.ParseLabel(tenant.Name),
		"application": platformK8s.ParseLabel(application.Name),
		"environment": platformK8s.ParseLabel(environment),
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	networkPolicy := &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: labels,
							},
						},
						AllowNetworkPolicyForMonitoring,
						AllowNetworkPolicyForSystemAPI,
					},
				},
			},
		},
	}
	return networkPolicy
}

// NewMongo create resource for mongo, including cronjob
// TODO Cronjob does not work locally without access to azure file storage (Test-Marka might help)
func NewMongo(environment string, tenant dolittleK8s.Tenant, application dolittleK8s.Application, settings MongoSettings) MongoResources {
	name := fmt.Sprintf("%s-mongo", strings.ToLower(environment))
	volumeMountName := fmt.Sprintf("%s-storage", name)

	storageClassName := "managed-premium"

	quantity, err := resource.ParseQuantity(settings.VolumeSize)
	if err != nil {
		log.Fatal(err)
	}

	namespace := fmt.Sprintf("application-%s", application.ID)

	labels := map[string]string{
		"tenant":         platformK8s.ParseLabel(tenant.Name),
		"application":    platformK8s.ParseLabel(application.Name),
		"environment":    platformK8s.ParseLabel(environment),
		"infrastructure": "Mongo",
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	statefulsetReplicas := int32(1)

	// Cronjob specific
	backupName := fmt.Sprintf("%s-backup", name)
	schedule := settings.CronJobSchedule
	successLimit := int32(1)
	failedLimit := int32(3)
	activeDeadlineSeconds := int64(600)
	mongoHost := fmt.Sprintf("%s.%s.svc.cluster.local:27017", name, namespace)
	// TODO Should we hard code it, instead of using Env variables?
	// TODO Should we make sure its lower case?
	archive := "/mnt/backup/$(APPLICATION)-$(ENVIRONMENT)-$(date +%Y-%m-%d_%H-%M-%S).gz.mongodump"
	shareName := settings.ShareName

	resource := MongoResources{
		Service: &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "corev1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Port: 27017,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Type(1),
							StrVal: "mongo",
						},
					}},
				Selector:  labels,
				ClusterIP: "None",
			},
		},
		StatefulSet: &appsv1.StatefulSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "StatefulSet",
				APIVersion: "apps/appsv1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &statefulsetReplicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      labels,
						Annotations: annotations,
					},
					Spec: corev1.PodSpec{Containers: []corev1.Container{
						{
							Name:  "mongo",
							Image: "dolittle/mongodb:4.2.2",
							Ports: []corev1.ContainerPort{
								{
									Name:          "mongo",
									ContainerPort: 27017,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      volumeMountName,
									MountPath: "/data/db",
								},
							},
						}}},
				},
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: volumeMountName,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.PersistentVolumeAccessMode("ReadWriteOnce"),
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{"storage": quantity},
							},
							StorageClassName: &storageClassName,
						},
					}},
				ServiceName:         name,
				PodManagementPolicy: appsv1.PodManagementPolicyType("OrderedReady"),
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.StatefulSetUpdateStrategyType("RollingUpdate"),
				},
			},
		},
		Cronjob: &v1beta1.CronJob{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CronJob",
				APIVersion: "batch/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels:      labels,
				Annotations: annotations,
				Name:        backupName,
				Namespace:   namespace,
			},
			Spec: v1beta1.CronJobSpec{
				Schedule:                   schedule,
				SuccessfulJobsHistoryLimit: &successLimit,
				FailedJobsHistoryLimit:     &failedLimit,
				JobTemplate: v1beta1.JobTemplateSpec{
					Spec: v1.JobSpec{
						ActiveDeadlineSeconds: &activeDeadlineSeconds,
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels:      labels,
								Annotations: annotations,
							},
							Spec: corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyNever,
								Containers: []corev1.Container{
									{
										Name:  "mongo-backup",
										Image: "dolittle/mongodb:4.2.2",
										Ports: []corev1.ContainerPort{
											{
												Name:          "mongo",
												ContainerPort: 27017,
											},
										},
										Command: []string{
											"bash",
											"-c",
										},
										Args: []string{
											fmt.Sprintf(`mongodump --host=%s --gzip --archive=%s`, mongoHost, archive),
										},
										Env: []corev1.EnvVar{
											{
												Name: "APPLICATION",
												ValueFrom: &corev1.EnvVarSource{
													FieldRef: &corev1.ObjectFieldSelector{
														FieldPath: "metadata.labels['application']",
													},
												},
											},
											{
												Name: "ENVIRONMENT",
												ValueFrom: &corev1.EnvVarSource{
													FieldRef: &corev1.ObjectFieldSelector{
														FieldPath: "metadata.labels['environment']",
													},
												},
											},
										},
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "backup-storage",
												MountPath: "/mnt/backup/",
												SubPath:   "mongo",
											},
										},
									},
								},
								Volumes: []corev1.Volume{
									{
										Name: "backup-storage",
										VolumeSource: corev1.VolumeSource{
											AzureFile: &corev1.AzureFileVolumeSource{
												SecretName: "storage-account-secret",
												ShareName:  shareName,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return resource
}

func NewLocalDevRoleBindingToDeveloper(tenant dolittleK8s.Tenant, application dolittleK8s.Application) *rbacv1.RoleBinding {
	namespace := fmt.Sprintf("application-%s", application.ID)
	labels := map[string]string{
		"tenant":      platformK8s.ParseLabel(tenant.Name),
		"application": platformK8s.ParseLabel(application.Name),
	}

	annotations := map[string]string{
		"dolittle.io/tenant-id":      tenant.ID,
		"dolittle.io/application-id": application.ID,
	}

	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/rbacv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "developer-local-dev",
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     "local-dev",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "developer",
		},
	}
}
