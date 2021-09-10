package rawdatalog

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type stanResources struct {
	configMap  *corev1.ConfigMap
	service    *corev1.Service
	statfulset *appsv1.StatefulSet
}

func createStanResources(namespace, environment string, labels, annotations labels.Set) stanResources {
	name := fmt.Sprintf("%s-stan", environment)
	natsName := fmt.Sprintf("%s-nats", environment)

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
		},
		Data: map[string]string{
			"stan.conf": fmt.Sprintf(`
				port: 4222
				http: 8222
			
				streaming {
					ns: "nats://%s:4222"
					id: stan
					store: MEMORY
				}
			`, natsName),
		},
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:  labels,
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       7777,
					TargetPort: intstr.FromString("metrics"),
				},
			},
		},
	}

	statefulsetReplicas := int32(1)

	statfulset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas:    &statefulsetReplicas,
			ServiceName: name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "stan",
							Image: "nats-streaming:0.22.0",
							Ports: []corev1.ContainerPort{
								{
									Name:          "monitor",
									ContainerPort: 8222,
								},
								{
									Name:          "metrics",
									ContainerPort: 7777,
								},
							},
							Args: []string{
								"-sc",
								"/etc/stan-config/stan.conf",
							},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/stan-config",
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromString("monitor"),
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
							},
						},
					},
				},
			},
		},
	}

	return stanResources{configMap, service, statfulset}
}
