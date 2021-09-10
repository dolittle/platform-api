package rawdatalog

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type natsResources struct {
	configMap  *corev1.ConfigMap
	service    *corev1.Service
	statfulset *appsv1.StatefulSet
}

func createNatsResources(namespace, environment string, labels, annotations labels.Set) natsResources {
	name := fmt.Sprintf("%s-nats", environment)

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
			"nats.conf": `
				pid_file: "/var/run/nats/nats.pid"
				http: 8222
			`,
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
					Name:       "client",
					Port:       4222,
					TargetPort: intstr.FromString("client"),
				},
				{
					Name:       "cluster",
					Port:       6222,
					TargetPort: intstr.FromString("cluster"),
				},
				{
					Name:       "monitor",
					Port:       8222,
					TargetPort: intstr.FromString("monitor"),
				},
				{
					Name:       "metrics",
					Port:       7777,
					TargetPort: intstr.FromString("metrics"),
				},
				{
					Name:       "leafnodes",
					Port:       7422,
					TargetPort: intstr.FromString("leafnodes"),
				},
				{
					Name: "gateways",
					Port: 7522,
				},
			},
		},
	}

	statefulsetReplicas := int32(1)
	statefulsetShareProcessNamespace := true
	statefulsetTerminationGracePeriodSeconds := int64(60)

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
						{
							Name: "pid",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					ShareProcessNamespace:         &statefulsetShareProcessNamespace,
					TerminationGracePeriodSeconds: &statefulsetTerminationGracePeriodSeconds,
					Containers: []corev1.Container{
						{
							Name:  "nats",
							Image: "nats:2.1.7-alpine3.11",
							Ports: []corev1.ContainerPort{
								{
									Name:          "client",
									ContainerPort: 4222,
								},
								{
									Name:          "cluster",
									ContainerPort: 6222,
								},
								{
									Name:          "monitor",
									ContainerPort: 8222,
								},
								{
									Name:          "metrics",
									ContainerPort: 7777,
								},
								{
									Name:          "leafnodes",
									ContainerPort: 7422,
								},
							},
							Command: []string{
								"nats-server",
								"--config",
								"/etc/nats-config/nats.conf",
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
								{
									Name:  "CLUSTER_ADVERTISE",
									Value: fmt.Sprintf("$(POD_NAME).%s.$(POD_NAMESPACE).svc.cluster.local", name),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/nats-config",
								},
								{
									Name:      "pid",
									MountPath: "/var/run/nats",
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
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromString("monitor"),
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
							},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "/nats-server -sl=ldm=/var/run/nats/nats.pid && /bin/sleep 60"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return natsResources{configMap, service, statfulset}
}
