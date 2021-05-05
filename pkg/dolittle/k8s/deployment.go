package k8s

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDeployment(microservice Microservice, headImage string, runtimeImage string) *appsv1.Deployment {
	labels := GetLabels(microservice)

	annotations := GetAnnotations(microservice)
	replicas := int32Ptr(1)

	deploymentName := fmt.Sprintf("%s-%s",
		microservice.Environment,
		microservice.Name,
	)

	dolittleConfigName := fmt.Sprintf("%s-%s-dolittle",
		microservice.Environment,
		microservice.Name,
	)
	tenantsConfigName := fmt.Sprintf("%s-tenants", microservice.Environment)

	configFilesName := fmt.Sprintf("%s-%s-config-files",
		microservice.Environment,
		microservice.Name,
	)

	configEnvVariablesName := fmt.Sprintf("%s-%s-env-variables",
		microservice.Environment,
		microservice.Name,
	)

	configSecretEnvVariablesName := fmt.Sprintf("%s-%s-secret-env-variables",
		microservice.Environment,
		microservice.Name,
	)

	deploymentName = strings.ToLower(deploymentName)
	dolittleConfigName = strings.ToLower(dolittleConfigName)
	tenantsConfigName = strings.ToLower(tenantsConfigName)
	configFilesName = strings.ToLower(configFilesName)
	configEnvVariablesName = strings.ToLower(configEnvVariablesName)
	configSecretEnvVariablesName = strings.ToLower(configSecretEnvVariablesName)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentName,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: apiv1.PodSpec{
					ImagePullSecrets: []apiv1.LocalObjectReference{
						{"acr"},
					},
					Containers: []apiv1.Container{
						{
							Name:  "head",
							Image: headImage,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							EnvFrom: []apiv1.EnvFromSource{
								{
									ConfigMapRef: &apiv1.ConfigMapEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: configEnvVariablesName,
										},
									},
								},
								{
									SecretRef: &apiv1.SecretEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: configSecretEnvVariablesName,
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									MountPath: "/app/.dolittle/resources.json",
									SubPath:   "resources.json",
									Name:      "dolittle-config",
								},
								{
									MountPath: "/app/data",
									Name:      "config-files",
								},
								{
									MountPath: "/app/.dolittle/event-horizons.json",
									SubPath:   "event-horizons.json",
									Name:      "dolittle-config",
								},
							},
						},
						Runtime(runtimeImage),
					},
					Volumes: []apiv1.Volume{
						{
							Name: "tenants-config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: tenantsConfigName,
									},
								},
							},
						},
						{
							Name: "dolittle-config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: dolittleConfigName,
									},
								},
							},
						},
						{
							Name: "config-files",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: configFilesName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

func Runtime(image string) apiv1.Container {
	return apiv1.Container{
		Name:  "runtime",
		Image: image,
		Ports: []apiv1.ContainerPort{
			{
				Name:          "runtime",
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: 50052,
			},
		},
		VolumeMounts: []apiv1.VolumeMount{
			{
				MountPath: "/app/.dolittle/tenants.json",
				SubPath:   "tenants.json",
				Name:      "tenants-config",
			},
			{
				MountPath: "/app/.dolittle/resources.json",
				SubPath:   "resources.json",
				Name:      "dolittle-config",
			},
			{
				MountPath: "/app/.dolittle/endpoints.json",
				SubPath:   "endpoints.json",
				Name:      "dolittle-config",
			},
			{
				MountPath: "/app/.dolittle/event-horizon-consents.json",
				SubPath:   "event-horizon-consents.json",
				Name:      "dolittle-config",
			},
			{
				MountPath: "/app/.dolittle/microservices.json",
				SubPath:   "microservices.json",
				Name:      "dolittle-config",
			},
			{
				MountPath: "/app/appsettings.json",
				SubPath:   "appsettings.json",
				Name:      "dolittle-config",
			},
		},
	}
}
func int32Ptr(i int32) *int32 { return &i }
