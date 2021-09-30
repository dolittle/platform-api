package rawdatalog

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type stanResources struct {
	configMap  *corev1.ConfigMap
	service    *corev1.Service
	statfulset *appsv1.StatefulSet
	backup     *v1beta1.CronJob
}

func createStanResources(namespace, environment string, labels, annotations labels.Set) stanResources {
	name := fmt.Sprintf("%s-stan", environment)
	natsName := fmt.Sprintf("%s-nats", environment)
	quantity, err := resource.ParseQuantity("8Gi")
	if err != nil {
		log.Fatal(err)
	}
	storageClassName := "managed-premium"
	storageName := fmt.Sprintf("%s-storage", name)
	storageDir := "datastore"

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
					store: file
					dir: %s
				}
			`, natsName, storageDir),
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
								{
									Name:      storageName,
									MountPath: storageDir,
									SubPath:   storageDir,
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
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: storageName,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{"storage": quantity},
						},
						StorageClassName: &storageClassName,
					},
				},
			},
		},
	}

	rand.Seed(time.Now().UnixNano())
	schedule := fmt.Sprintf("%v * * * *", rand.Intn(59))
	successfulJobsLimit := int32(1)
	failedJobsLimit := int32(3)
	activeDeadlineSeconds := int64(600)

	backup := &v1beta1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1beta1",
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-backup", storageName),
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: v1beta1.CronJobSpec{
			Schedule:                   schedule,
			SuccessfulJobsHistoryLimit: &successfulJobsLimit,
			FailedJobsHistoryLimit:     &failedJobsLimit,
			JobTemplate: v1beta1.JobTemplateSpec{
				Spec: v1.JobSpec{
					ActiveDeadlineSeconds: &activeDeadlineSeconds,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: annotations,
							Labels:      labels,
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "Never",
							Containers: []corev1.Container{
								{
									Name:  "stan-backup",
									Image: "alpine:3.14",
									Command: []string{
										"/bin/bash",
										"-c",
										"--",
									},
									Args: []string{
										fmt.Sprintf("tar -cvzf /mnt/backup/%s-$(date +%%Y-%%m-%%d_%%H-%%M-%%S).tar.gz %s", environment, storageDir),
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											MountPath: "/mnt/backup/",
											SubPath:   "stan",
											Name:      "backup-storage",
										},
										{
											Name:      storageName,
											MountPath: storageDir,
											SubPath:   storageDir,
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
											ShareName: fmt.Sprintf("%s-%s-nats-backup",
												strings.ToLower(labels["application"]),
												environment),
										},
									},
								},
								{
									Name: storageName,
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: storageName,
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

	return stanResources{configMap, service, statfulset, backup}
}
