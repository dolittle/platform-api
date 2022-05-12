package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/dolittle/platform-api/pkg/dolittle/k8s"
)

var _ = Describe("Deployment", func() {
	Describe("when creating a NewDeployment", func() {
		var (
			microservice Microservice
			headImage    string
			runtimeImage string
			deployment   *appsv1.Deployment
		)

		BeforeEach(func() {
			microservice = Microservice{
				ID:          "c974b165-38d7-4745-9c62-f78fa615682a",
				Name:        "LeliaKim",
				Environment: "AndreJensen",
				Tenant: Tenant{
					ID:   "4acf6e5b-6fb2-4a29-8073-3a79707ab558",
					Name: "JeanetteJohnston",
				},
				Application: Application{
					ID:   "cc142a0d-deac-4974-ada9-de6e21337dca",
					Name: "AlejandroRiley",
				},
			}
			headImage = "feeeea48-16cc-4cad-8cac-d00e65969747.repository.io/StevenPerkins:1.0.0-alpha.1"
			runtimeImage = "dolittle/runtime:160.1.0"

			deployment = NewDeployment(microservice, headImage, runtimeImage)
		})

		It("should create a deployment with the correct ApiVersion", func() {
			Expect(deployment.APIVersion).To(Equal("apps/v1"))
		})
		It("should create a deployment with the correct Kind", func() {
			Expect(deployment.Kind).To(Equal("Deployment"))
		})
		It("should create a deployment with the correct tenant-id annotation", func() {
			Expect(deployment.Annotations["dolittle.io/tenant-id"]).To(Equal(microservice.Tenant.ID))
		})
		It("should create a deployment with the correct application-id annotation", func() {
			Expect(deployment.Annotations["dolittle.io/application-id"]).To(Equal(microservice.Application.ID))
		})
		It("should create a deployment with the correct microservice-id annotation", func() {
			Expect(deployment.Annotations["dolittle.io/microservice-id"]).To(Equal(microservice.ID))
		})
		It("should create a deployment with the correct tenant label", func() {
			Expect(deployment.Labels["tenant"]).To(Equal(microservice.Tenant.Name))
		})
		It("should create a deployment with the correct application label", func() {
			Expect(deployment.Labels["application"]).To(Equal(microservice.Application.Name))
		})
		It("should create a deployment with the correct environment label", func() {
			Expect(deployment.Labels["environment"]).To(Equal(microservice.Environment))
		})
		It("should create a deployment with the correct microservice label", func() {
			Expect(deployment.Labels["microservice"]).To(Equal(microservice.Name))
		})
		It("should create a deployment with the correct tenant label selector", func() {
			Expect(deployment.Spec.Selector.MatchLabels["tenant"]).To(Equal(microservice.Tenant.Name))
		})
		It("should create a deployment with the correct tenant application selector", func() {
			Expect(deployment.Spec.Selector.MatchLabels["application"]).To(Equal(microservice.Application.Name))
		})
		It("should create a deployment with the correct tenant environment selector", func() {
			Expect(deployment.Spec.Selector.MatchLabels["environment"]).To(Equal(microservice.Environment))
		})
		It("should create a deployment with the correct tenant microservice selector", func() {
			Expect(deployment.Spec.Selector.MatchLabels["microservice"]).To(Equal(microservice.Name))
		})
		It("should create a pod template with the correct tenant-id annotation", func() {
			Expect(deployment.Spec.Template.Annotations["dolittle.io/tenant-id"]).To(Equal(microservice.Tenant.ID))
		})
		It("should create a pod template with the correct application-id annotation", func() {
			Expect(deployment.Spec.Template.Annotations["dolittle.io/application-id"]).To(Equal(microservice.Application.ID))
		})
		It("should create a pod template with the correct microservice-id annotation", func() {
			Expect(deployment.Spec.Template.Annotations["dolittle.io/microservice-id"]).To(Equal(microservice.ID))
		})
		It("should create a pod template with the correct tenant label", func() {
			Expect(deployment.Spec.Template.Labels["tenant"]).To(Equal(microservice.Tenant.Name))
		})
		It("should create a pod template with the correct application label", func() {
			Expect(deployment.Spec.Template.Labels["application"]).To(Equal(microservice.Application.Name))
		})
		It("should create a pod template with the correct environment label", func() {
			Expect(deployment.Spec.Template.Labels["environment"]).To(Equal(microservice.Environment))
		})
		It("should create a pod template with the correct microservice label", func() {
			Expect(deployment.Spec.Template.Labels["microservice"]).To(Equal(microservice.Name))
		})
		It("should create a pod template with the 'acr' imagePullSecrets", func() {
			Expect(deployment.Spec.Template.Spec.ImagePullSecrets[0].Name).To(Equal("acr"))
		})
		It("should create a container with named 'head'", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(Equal("head"))
		})
		It("should create a head container with the correct image", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(headImage))
		})
		It("should create a head container with port 80 exposed", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(int32(80)))
		})
		It("should create a head container with port 80 named 'http'", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports[0].Name).To(Equal("http"))
		})
		It("should create a head container with environmental variables from the correct configmap", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].EnvFrom[0].ConfigMapRef.Name).To(Equal("andrejensen-leliakim-env-variables"))
		})
		It("should create a head container with environmental variables from the correct configmap", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].EnvFrom[1].SecretRef.Name).To(Equal("andrejensen-leliakim-secret-env-variables"))
		})
		It("should create a head container with 'tenants.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/app/.dolittle/tenants.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].SubPath).To(Equal("tenants.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(Equal("tenants-config"))
		})
		It("should create a head container with 'resources.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath).To(Equal("/app/.dolittle/resources.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1].SubPath).To(Equal("resources.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name).To(Equal("dolittle-config"))
		})
		It("should create a head container with 'clients.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[2].MountPath).To(Equal("/app/.dolittle/clients.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[2].SubPath).To(Equal("clients.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[2].Name).To(Equal("dolittle-config"))
		})
		It("should create a head container with 'event-horizons.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[3].MountPath).To(Equal("/app/.dolittle/event-horizons.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[3].SubPath).To(Equal("event-horizons.json"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[3].Name).To(Equal("dolittle-config"))
		})
		It("should create a head container with config files mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[4].MountPath).To(Equal("/app/data"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[4].Name).To(Equal("config-files"))
		})

		It("should create a head container with resource requests and limits", func() {
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).To(Equal("25m"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).To(Equal("256Mi"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()).To(Equal("2"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).To(Equal("1Gi"))
		})

		It("should create a container named 'runtime'", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].Name).To(Equal("runtime"))
		})
		It("should create a runtime container with the correct image", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].Image).To(Equal(runtimeImage))
		})
		It("should create a runtime container with port 50052 exposed", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].Ports[0].ContainerPort).To(Equal(int32(50052)))
		})
		It("should create a runtime container with port 50052 named 'runtime'", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].Ports[0].Name).To(Equal("runtime"))
		})
		It("should create a runtime container with 'tenants.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[0].MountPath).To(Equal("/app/.dolittle/tenants.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[0].SubPath).To(Equal("tenants.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name).To(Equal("tenants-config"))
		})
		It("should create a runtime container with 'resources.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[1].MountPath).To(Equal("/app/.dolittle/resources.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[1].SubPath).To(Equal("resources.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[1].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'endpoints.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[2].MountPath).To(Equal("/app/.dolittle/endpoints.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[2].SubPath).To(Equal("endpoints.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[2].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'event-horizon-consents.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[3].MountPath).To(Equal("/app/.dolittle/event-horizon-consents.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[3].SubPath).To(Equal("event-horizon-consents.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[3].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'microservices.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[4].MountPath).To(Equal("/app/.dolittle/microservices.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[4].SubPath).To(Equal("microservices.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[4].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'platform.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[5].MountPath).To(Equal("/app/.dolittle/platform.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[5].SubPath).To(Equal("platform.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[5].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'appsettings.json' mounted", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[6].MountPath).To(Equal("/app/appsettings.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[6].SubPath).To(Equal("appsettings.json"))
			Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts[6].Name).To(Equal("dolittle-config"))
		})

		It("should create a runtime container with resource limits", func() {
			Expect(deployment.Spec.Template.Spec.Containers[1].Resources.Requests.Cpu().String()).To(Equal("25m"))
			Expect(deployment.Spec.Template.Spec.Containers[1].Resources.Requests.Memory().String()).To(Equal("256Mi"))
			Expect(deployment.Spec.Template.Spec.Containers[1].Resources.Limits.Cpu().String()).To(Equal("2"))
			Expect(deployment.Spec.Template.Spec.Containers[1].Resources.Limits.Memory().String()).To(Equal("1Gi"))
		})

		It("should create a pod template with the 'tenants-config' volume", func() {
			Expect(deployment.Spec.Template.Spec.Volumes[0].Name).To(Equal("tenants-config"))
			Expect(deployment.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("andrejensen-tenants"))
		})
		It("should create a pod template with the 'dolittle-config' volume", func() {
			Expect(deployment.Spec.Template.Spec.Volumes[1].Name).To(Equal("dolittle-config"))
			Expect(deployment.Spec.Template.Spec.Volumes[1].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("andrejensen-leliakim-dolittle"))
		})
		It("should create a pod template with the 'config-files' volume", func() {
			Expect(deployment.Spec.Template.Spec.Volumes[2].Name).To(Equal("config-files"))
			Expect(deployment.Spec.Template.Spec.Volumes[2].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("andrejensen-leliakim-config-files"))
		})
	})

	Describe("when creating a Runtime for a Prod environment", func() {
		var (
			runtimeImage string
			environment  string
			container    corev1.Container
		)

		BeforeEach(func() {
			runtimeImage = "dolittle/runtime:160.1.0"
			environment = "Prod"

			container = Runtime(runtimeImage, environment)
		})

		It("should create a container named 'runtime'", func() {
			Expect(container.Name).To(Equal("runtime"))
		})
		It("should create a runtime container with the correct image", func() {
			Expect(container.Image).To(Equal(runtimeImage))
		})
		It("should create a runtime container with port 9700 exposed", func() {
			Expect(container.Ports[1].ContainerPort).To(Equal(int32(9700)))
		})
		It("should create a runtime container with port 9700 named 'runtime-metrics'", func() {
			Expect(container.Ports[1].Name).To(Equal("runtime-metrics"))
		})

		It("should create a runtime container with port 50052 exposed", func() {
			Expect(container.Ports[0].ContainerPort).To(Equal(int32(50052)))
		})
		It("should create a runtime container with port 50052 named 'runtime'", func() {
			Expect(container.Ports[0].Name).To(Equal("runtime"))
		})

		It("should create a runtime container with 'tenants.json' mounted", func() {
			Expect(container.VolumeMounts[0].MountPath).To(Equal("/app/.dolittle/tenants.json"))
			Expect(container.VolumeMounts[0].SubPath).To(Equal("tenants.json"))
			Expect(container.VolumeMounts[0].Name).To(Equal("tenants-config"))
		})
		It("should create a runtime container with 'resources.json' mounted", func() {
			Expect(container.VolumeMounts[1].MountPath).To(Equal("/app/.dolittle/resources.json"))
			Expect(container.VolumeMounts[1].SubPath).To(Equal("resources.json"))
			Expect(container.VolumeMounts[1].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'endpoints.json' mounted", func() {
			Expect(container.VolumeMounts[2].MountPath).To(Equal("/app/.dolittle/endpoints.json"))
			Expect(container.VolumeMounts[2].SubPath).To(Equal("endpoints.json"))
			Expect(container.VolumeMounts[2].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'event-horizon-consents.json' mounted", func() {
			Expect(container.VolumeMounts[3].MountPath).To(Equal("/app/.dolittle/event-horizon-consents.json"))
			Expect(container.VolumeMounts[3].SubPath).To(Equal("event-horizon-consents.json"))
			Expect(container.VolumeMounts[3].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'microservices.json' mounted", func() {
			Expect(container.VolumeMounts[4].MountPath).To(Equal("/app/.dolittle/microservices.json"))
			Expect(container.VolumeMounts[4].SubPath).To(Equal("microservices.json"))
			Expect(container.VolumeMounts[4].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'platform.json' mounted", func() {
			Expect(container.VolumeMounts[5].MountPath).To(Equal("/app/.dolittle/platform.json"))
			Expect(container.VolumeMounts[5].SubPath).To(Equal("platform.json"))
			Expect(container.VolumeMounts[5].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with 'appsettings.json' mounted", func() {
			Expect(container.VolumeMounts[6].MountPath).To(Equal("/app/appsettings.json"))
			Expect(container.VolumeMounts[6].SubPath).To(Equal("appsettings.json"))
			Expect(container.VolumeMounts[6].Name).To(Equal("dolittle-config"))
		})
		It("should create a runtime container with resource limits", func() {
			Expect(container.Resources.Requests.Cpu().String()).To(Equal("50m"))
			Expect(container.Resources.Requests.Memory().String()).To(Equal("256Mi"))
			Expect(container.Resources.Limits.Cpu().String()).To(Equal("2"))
			Expect(container.Resources.Limits.Memory().String()).To(Equal("1Gi"))
		})
	})
})
