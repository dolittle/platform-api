package k8s

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

	// TODO in the future, this could be linked to the customers subscription
	// Or some way to let them override it as a premium feature
	headResourceLimit, err := resource.ParseQuantity("500Mi")
	if err != nil {
		panic(err)
	}

	headResourceRequest, err := resource.ParseQuantity("250Mi")
	if err != nil {
		panic(err)
	}

	containers := []apiv1.Container{
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
			Resources: apiv1.ResourceRequirements{
				Limits: apiv1.ResourceList{
					apiv1.ResourceMemory: headResourceLimit,
				},
				Requests: apiv1.ResourceList{
					apiv1.ResourceMemory: headResourceRequest,
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
					MountPath: "/app/.dolittle/clients.json",
					SubPath:   "clients.json",
					Name:      "dolittle-config",
				},
				{
					MountPath: "/app/.dolittle/event-horizons.json",
					SubPath:   "event-horizons.json",
					Name:      "dolittle-config",
				},
				{
					MountPath: "/app/data",
					Name:      "config-files",
				},
			},
		},
	}
	if runtimeImage != "none" {
		containers = append(containers, Runtime(runtimeImage))
	}

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
						{
							Name: "acr",
						},
					},
					Containers: containers,
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
	limit, err := resource.ParseQuantity("1Gi")
	if err != nil {
		panic(err)
	}
	request, err := resource.ParseQuantity("250Mi")
	if err != nil {
		panic(err)
	}
	return apiv1.Container{
		Name:  "runtime",
		Image: image,
		Ports: []apiv1.ContainerPort{
			{
				Name:          "runtime",
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: 50052,
			},
			{
				Name:          "runtime-metrics",
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: 9700,
			},
		},
		Resources: apiv1.ResourceRequirements{
			Limits: apiv1.ResourceList{
				apiv1.ResourceMemory: limit,
			},
			Requests: apiv1.ResourceList{
				apiv1.ResourceMemory: request,
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
				MountPath: "/app/.dolittle/platform.json",
				SubPath:   "platform.json",
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
