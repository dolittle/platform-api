package k8s

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewService(microservice Microservice) *corev1.Service {
	serviceName := fmt.Sprintf("%s-%s",
		microservice.Environment,
		microservice.Name,
	)

	labels := GetLabels(microservice)

	annotations := map[string]string{
		"dolittle.io/tenant-id":       microservice.Tenant.ID,
		"dolittle.io/application-id":  microservice.Application.ID,
		"dolittle.io/microservice-id": microservice.ID,
	}

	serviceName = strings.ToLower(serviceName)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceName,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "http",
					},
					Protocol: corev1.ProtocolTCP,
					Port:     80,
				},
				{
					Name: "runtime",
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "runtime",
					},
					Protocol: corev1.ProtocolTCP,
					Port:     50052,
				},
			},
		},
	}
}
