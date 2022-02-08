package rawdatalog_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	. "github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
)

var _ = Describe("Repo", func() {
	var (
		clientSet      *fake.Clientset
		config         *rest.Config
		k8sRepo        platformK8s.K8sRepo
		logger         *logrus.Logger
		rawDataLogRepo RawDataLogIngestorRepo
		isProduction   bool
	)

	BeforeEach(func() {
		logger, _ = logrusTest.NewNullLogger()
		clientSet = fake.NewSimpleClientset()
		config = &rest.Config{}
		k8sRepo = platformK8s.NewK8sRepo(clientSet, config, logger)
		isProduction = false
		rawDataLogRepo = NewRawDataLogIngestorRepo(isProduction, k8sRepo, clientSet, logger)
	})

	Describe("when creating RawDataLog", func() {
		var (
			namespace   string
			tenant      k8s.Tenant
			application k8s.Application
			input       platform.HttpInputRawDataLogIngestorInfo
			err         error
		)

		BeforeEach(func() {
			namespace = "application-6db1278e-da39-481a-8474-e0ef6bdc2f6e"
			tenant = k8s.Tenant{
				Name: "LydiaBall",
				ID:   "c6c72dab-a770-47d5-b85d-2777d2ac0922",
			}
			application = k8s.Application{
				Name: "CordeliaChavez",
				ID:   "6db1278e-da39-481a-8474-e0ef6bdc2f6e",
			}
			input = platform.HttpInputRawDataLogIngestorInfo{
				MicroserviceBase: platform.MicroserviceBase{
					Environment: "LoisMay",
					Name:        "ErnestBush",
					Dolittle: platform.HttpInputDolittle{
						ApplicationID:  application.ID,
						TenantID:       tenant.ID,
						MicroserviceID: "b9a9211e-f118-4ea0-9eb9-8d0d8f33c753",
					},
					Kind: platform.MicroserviceKindRawDataLogIngestor,
				},
				Extra: platform.HttpInputRawDataLogIngestorExtra{
					Ingress: platform.HttpInputSimpleIngress{
						Path:     "/api/not-webhooks-just-to-be-sure",
						Pathtype: "SpecialTypeNotActuallySupported",
					},
				},
			}
		})

		Context("for an application that does not have any ingresses", func() {
			BeforeEach(func() {
				customerTenants := make([]platform.CustomerTenantInfo, 0)
				err = rawDataLogRepo.Create(namespace, tenant, application, customerTenants, input)
			})

			It("should fail with an error", func() {
				Expect(err).ToNot(BeNil())
			})
			It("should not create any resources", func() {
				Expect(getCreateActions(clientSet)).To(BeEmpty())
			})
		})

		Context("and an application exists but no other resources", func() {
			var (
				natsConfigMap             *corev1.ConfigMap
				natsService               *corev1.Service
				natsStatefulSet           *appsv1.StatefulSet
				stanConfigMap             *corev1.ConfigMap
				stanService               *corev1.Service
				stanStatefulSet           *appsv1.StatefulSet
				rawDataLogIngestorIngress *netv1.Ingress
				rawDataLogDeployment      *appsv1.Deployment
				customerTenantID          string
			)

			BeforeEach(func() {
				customerTenantID = "f4679b71-1215-4a60-8483-53b0d5f2bb47"
				customerTenants := []platform.CustomerTenantInfo{
					{
						CustomerTenantID: customerTenantID,
						Hosts: []platform.CustomerTenantHost{
							{
								Host:       "some-fancy.domain.name",
								SecretName: "some-fancy-certificate",
							},
						},
					},
				}
				err = rawDataLogRepo.Create(namespace, tenant, application, customerTenants, input)
			})

			// NATS ConfigMap
			It("should create a configmap for nats named 'loismay-nats'", func() {
				object := getCreatedObject(clientSet, "ConfigMap", "loismay-nats")
				Expect(object).ToNot(BeNil())
				natsConfigMap = object.(*corev1.ConfigMap)
			})
			It("should create a configmap for nats with the correct ApiVersion", func() {
				Expect(natsConfigMap.APIVersion).To(Equal("v1"))
			})
			It("should create a configmap for nats with the correct Kind", func() {
				Expect(natsConfigMap.Kind).To(Equal("ConfigMap"))
			})
			It("should create a configmap for nats with the correct tenant-id annotation", func() {
				Expect(natsConfigMap.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a configmap for nats with the correct application-id annotation", func() {
				Expect(natsConfigMap.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a configmap for nats with the correct tenant label", func() {
				Expect(natsConfigMap.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a configmap for nats with the correct application label", func() {
				Expect(natsConfigMap.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a configmap for nats with the correct environment label", func() {
				Expect(natsConfigMap.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a configmap for nats with the correct infrastructure label", func() {
				Expect(natsConfigMap.Labels["infrastructure"]).To(Equal("Nats"))
			})
			It("should create a configmap for nats without a microservice label", func() {
				Expect(natsConfigMap.Labels["microservice"]).To(Equal(""))
			})
			It("should create a configmap for nats with 'nats.conf'", func() {
				Expect(natsConfigMap.Data["nats.conf"]).To(Equal(`
				pid_file: "/var/run/nats/nats.pid"
				http: 8222
			`))
			})

			// NATS Service
			It("should create a service for nats named 'loismay-nats'", func() {
				object := getCreatedObject(clientSet, "Service", "loismay-nats")
				Expect(object).ToNot(BeNil())
				natsService = object.(*corev1.Service)
			})
			It("should create a service for nats with the correct ApiVersion", func() {
				Expect(natsService.APIVersion).To(Equal("v1"))
			})
			It("should create a service for nats with the correct Kind", func() {
				Expect(natsService.Kind).To(Equal("Service"))
			})
			It("should create a service for nats that is headless", func() {
				Expect(natsService.Spec.ClusterIP).To(Equal("None"))
			})
			It("should create a service for nats with the correct tenant-id annotation", func() {
				Expect(natsService.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a service for nats with the correct application-id annotation", func() {
				Expect(natsService.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a service for nats with the correct tenant label", func() {
				Expect(natsService.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a service for nats with the correct application label", func() {
				Expect(natsService.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a service for nats with the correct environment label", func() {
				Expect(natsService.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a service for nats with the correct infrastructure label", func() {
				Expect(natsService.Labels["infrastructure"]).To(Equal("Nats"))
			})
			It("should create a service for nats without a microservice label", func() {
				Expect(natsService.Labels["microservice"]).To(Equal(""))
			})
			It("should create a service for nats with the correct tenant label selector", func() {
				Expect(natsService.Spec.Selector["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a service for nats with the correct application label selector", func() {
				Expect(natsService.Spec.Selector["application"]).To(Equal(application.Name))
			})
			It("should create a service for nats with the correct environment label selector", func() {
				Expect(natsService.Spec.Selector["environment"]).To(Equal(input.Environment))
			})
			It("should create a service for nats with the correct infrastructure label selector", func() {
				Expect(natsService.Spec.Selector["infrastructure"]).To(Equal("Nats"))
			})
			It("should create a service for nats without a microservice label selector", func() {
				Expect(natsService.Spec.Selector["microservice"]).To(Equal(""))
			})
			It("should create a service for nats with the 'client' port exposed", func() {
				Expect(natsService.Spec.Ports[0].Name).To(Equal("client"))
				Expect(natsService.Spec.Ports[0].Port).To(Equal(int32(4222)))
				Expect(natsService.Spec.Ports[0].TargetPort.StrVal).To(Equal("client"))
			})
			It("should create a service for nats with the 'cluster' port exposed", func() {
				Expect(natsService.Spec.Ports[1].Name).To(Equal("cluster"))
				Expect(natsService.Spec.Ports[1].Port).To(Equal(int32(6222)))
				Expect(natsService.Spec.Ports[1].TargetPort.StrVal).To(Equal("cluster"))
			})
			It("should create a service for nats with the 'monitor' port exposed", func() {
				Expect(natsService.Spec.Ports[2].Name).To(Equal("monitor"))
				Expect(natsService.Spec.Ports[2].Port).To(Equal(int32(8222)))
				Expect(natsService.Spec.Ports[2].TargetPort.StrVal).To(Equal("monitor"))
			})
			It("should create a service for nats with the 'metrics' port exposed", func() {
				Expect(natsService.Spec.Ports[3].Name).To(Equal("metrics"))
				Expect(natsService.Spec.Ports[3].Port).To(Equal(int32(7777)))
				Expect(natsService.Spec.Ports[3].TargetPort.StrVal).To(Equal("metrics"))
			})
			It("should create a service for nats with the 'leafnodes' port exposed", func() {
				Expect(natsService.Spec.Ports[4].Name).To(Equal("leafnodes"))
				Expect(natsService.Spec.Ports[4].Port).To(Equal(int32(7422)))
				Expect(natsService.Spec.Ports[4].TargetPort.StrVal).To(Equal("leafnodes"))
			})
			It("should create a service for nats with the 'gateways' port exposed", func() {
				Expect(natsService.Spec.Ports[5].Name).To(Equal("gateways"))
				Expect(natsService.Spec.Ports[5].Port).To(Equal(int32(7522)))
			})

			// NATS StatefulSet
			It("should create a statefulset for nats named 'loismay-nats'", func() {
				object := getCreatedObject(clientSet, "StatefulSet", "loismay-nats")
				Expect(object).ToNot(BeNil())
				natsStatefulSet = object.(*appsv1.StatefulSet)
			})
			It("should create a statefulset for nats with the correct ApiVersion", func() {
				Expect(natsStatefulSet.APIVersion).To(Equal("apps/v1"))
			})
			It("should create a statefulset for nats with the correct Kind", func() {
				Expect(natsStatefulSet.Kind).To(Equal("StatefulSet"))
			})
			It("should create a statefulset for nats with the correct tenant-id annotation", func() {
				Expect(natsStatefulSet.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a statefulset for nats with the correct application-id annotation", func() {
				Expect(natsStatefulSet.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a statefulset for nats with the correct tenant label", func() {
				Expect(natsStatefulSet.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a statefulset for nats with the correct application label", func() {
				Expect(natsStatefulSet.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a statefulset for nats with the correct environment label", func() {
				Expect(natsStatefulSet.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a statefulset for nats with the correct infrastructure label", func() {
				Expect(natsStatefulSet.Labels["infrastructure"]).To(Equal("Nats"))
			})
			It("should create a statefulset for nats without a microservice label", func() {
				Expect(natsStatefulSet.Labels["microservice"]).To(Equal(""))
			})
			It("should create a statefulset for nats with the correct tenant label selector", func() {
				Expect(natsStatefulSet.Spec.Selector.MatchLabels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a statefulset for nats with the correct application label selector", func() {
				Expect(natsStatefulSet.Spec.Selector.MatchLabels["application"]).To(Equal(application.Name))
			})
			It("should create a statefulset for nats with the correct environment label selector", func() {
				Expect(natsStatefulSet.Spec.Selector.MatchLabels["environment"]).To(Equal(input.Environment))
			})
			It("should create a statefulset for nats with the correct environment label selector", func() {
				Expect(natsStatefulSet.Spec.Selector.MatchLabels["infrastructure"]).To(Equal("Nats"))
			})
			It("should create a statefulset for nats without a microservice label selector", func() {
				Expect(natsStatefulSet.Spec.Selector.MatchLabels["microservice"]).To(Equal(""))
			})
			It("should create a statefulset for nats with one replica", func() {
				Expect(*natsStatefulSet.Spec.Replicas).To(Equal(int32(1)))
			})
			It("should create a pod template for nats with the correct tenant-id annotation", func() {
				Expect(natsStatefulSet.Spec.Template.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a pod template for nats with the correct application-id annotation", func() {
				Expect(natsStatefulSet.Spec.Template.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a pod template for nats with the correct tenant label", func() {
				Expect(natsStatefulSet.Spec.Template.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a pod template for nats with the correct application label", func() {
				Expect(natsStatefulSet.Spec.Template.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a pod template for nats with the correct environment label", func() {
				Expect(natsStatefulSet.Spec.Template.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a pod template for nats with the correct infrastructure label", func() {
				Expect(natsStatefulSet.Spec.Template.Labels["infrastructure"]).To(Equal("Nats"))
			})
			It("should create a pod template for nats without a microservice label", func() {
				Expect(natsStatefulSet.Spec.Template.Labels["microservice"]).To(Equal(""))
			})
			It("should create a pod template for nats that shares the process namespace", func() {
				Expect(*natsStatefulSet.Spec.Template.Spec.ShareProcessNamespace).To(Equal(true))
			})
			It("should create a pod template for nats with the nats config map as a volume", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
				Expect(natsStatefulSet.Spec.Template.Spec.Volumes[0].ConfigMap.Name).To(Equal("loismay-nats"))
			})
			It("should create a pod template for nats with a pid volume", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Volumes[1].Name).To(Equal("pid"))
				Expect(natsStatefulSet.Spec.Template.Spec.Volumes[1].EmptyDir).ToNot(Equal(""))
			})
			It("should create a container for nats named 'nats'", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Name).To(Equal("nats"))
			})
			It("should create a container for nats with the correct image", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Image).To(Equal("nats:2.1.7-alpine3.11"))
			})
			It("should create a container for nats with the correct command", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Command[0]).To(Equal("nats-server"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Command[1]).To(Equal("--config"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Command[2]).To(Equal("/etc/nats-config/nats.conf"))
			})
			It("should create a container for nats with the 'client' port exposed", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(int32(4222)))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[0].Name).To(Equal("client"))
			})
			It("should create a container for nats with the 'cluster' port exposed", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[1].ContainerPort).To(Equal(int32(6222)))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[1].Name).To(Equal("cluster"))
			})
			It("should create a container for nats with the 'monitor' port exposed", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[2].ContainerPort).To(Equal(int32(8222)))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[2].Name).To(Equal("monitor"))
			})
			It("should create a container for nats with the 'metrics' port exposed", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[3].ContainerPort).To(Equal(int32(7777)))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[3].Name).To(Equal("metrics"))
			})
			It("should create a container for nats with the 'leafnodes' port exposed", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[4].ContainerPort).To(Equal(int32(7422)))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Ports[4].Name).To(Equal("leafnodes"))
			})
			It("should create a container for nats with the 'POD_NAME' environmental variable set", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("POD_NAME"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Env[0].ValueFrom.FieldRef.FieldPath).To(Equal("metadata.name"))
			})
			It("should create a container for nats with the 'POD_NAMESPACE' environmental variable set", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Env[1].Name).To(Equal("POD_NAMESPACE"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Env[1].ValueFrom.FieldRef.FieldPath).To(Equal("metadata.namespace"))
			})
			It("should create a container for nats with the 'CLUSTER_ADVERTISE' environmental variable set", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Env[2].Name).To(Equal("CLUSTER_ADVERTISE"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Env[2].Value).To(Equal("$(POD_NAME).loismay-nats.$(POD_NAMESPACE).svc.cluster.local"))
			})
			It("should create a container for nats with '/etc/nats-config' mounted", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/etc/nats-config"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(Equal("config-volume"))
			})
			It("should create a container for nats with '/var/run/nats' mounted", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath).To(Equal("/var/run/nats"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name).To(Equal("pid"))
			})
			It("should create a container for nats with a liveness probe", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).To(Equal("/"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.StrVal).To(Equal("monitor"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds).To(Equal(int32(10)))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds).To(Equal(int32(5)))
			})
			It("should create a container for nats with a readiness probe", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).To(Equal("/"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.StrVal).To(Equal("monitor"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.InitialDelaySeconds).To(Equal(int32(10)))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.TimeoutSeconds).To(Equal(int32(5)))
			})
			It("should create a container for nats with a prestop lifecycle command to shut it down gracefully", func() {
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.Exec.Command[0]).To(Equal("/bin/sh"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.Exec.Command[1]).To(Equal("-c"))
				Expect(natsStatefulSet.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.Exec.Command[2]).To(Equal("/nats-server -sl=ldm=/var/run/nats/nats.pid && /bin/sleep 60"))
			})

			// STAN ConfigMap
			It("should create a configmap for stan named 'loismay-stan'", func() {
				object := getCreatedObject(clientSet, "ConfigMap", "loismay-stan")
				Expect(object).ToNot(BeNil())
				stanConfigMap = object.(*corev1.ConfigMap)
			})
			It("should create a configmap for stan with the correct ApiVersion", func() {
				Expect(stanConfigMap.APIVersion).To(Equal("v1"))
			})
			It("should create a configmap for stan with the correct Kind", func() {
				Expect(stanConfigMap.Kind).To(Equal("ConfigMap"))
			})
			It("should create a configmap for stan with the correct tenant-id annotation", func() {
				Expect(stanConfigMap.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a configmap for stan with the correct application-id annotation", func() {
				Expect(stanConfigMap.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a configmap for stan with the correct tenant label", func() {
				Expect(stanConfigMap.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a configmap for stan with the correct application label", func() {
				Expect(stanConfigMap.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a configmap for stan with the correct environment label", func() {
				Expect(stanConfigMap.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a configmap for stan with the correct infrastructure label", func() {
				Expect(stanConfigMap.Labels["infrastructure"]).To(Equal("Stan"))
			})
			It("should create a configmap for stan without a microservice label", func() {
				Expect(stanConfigMap.Labels["microservice"]).To(Equal(""))
			})
			It("should create a configmap for stan with 'stan.conf'", func() {
				Expect(stanConfigMap.Data["stan.conf"]).To(Equal(`
				port: 4222
				http: 8222
			
				streaming {
					ns: "nats://loismay-nats:4222"
					id: stan
					store: file
					dir: datastore
				}
			`))
			})

			// STAN Service
			It("should create a service for stan named 'loismay-nats'", func() {
				object := getCreatedObject(clientSet, "Service", "loismay-stan")
				Expect(object).ToNot(BeNil())
				stanService = object.(*corev1.Service)
			})
			It("should create a service for stan with the correct ApiVersion", func() {
				Expect(stanService.APIVersion).To(Equal("v1"))
			})
			It("should create a service for stan with the correct Kind", func() {
				Expect(stanService.Kind).To(Equal("Service"))
			})
			It("should create a service for stan that is headless", func() {
				Expect(stanService.Spec.ClusterIP).To(Equal("None"))
			})
			It("should create a service for stan with the correct tenant-id annotation", func() {
				Expect(stanService.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a service for stan with the correct application-id annotation", func() {
				Expect(stanService.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a service for stan with the correct tenant label", func() {
				Expect(stanService.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a service for stan with the correct application label", func() {
				Expect(stanService.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a service for stan with the correct environment label", func() {
				Expect(stanService.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a service for stan with the correct infrastructure label", func() {
				Expect(stanService.Labels["infrastructure"]).To(Equal("Stan"))
			})
			It("should create a service for stan without a microservice label", func() {
				Expect(stanService.Labels["microservice"]).To(Equal(""))
			})
			It("should create a service for stan with the correct tenant label selector", func() {
				Expect(stanService.Spec.Selector["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a service for stan with the correct application label selector", func() {
				Expect(stanService.Spec.Selector["application"]).To(Equal(application.Name))
			})
			It("should create a service for stan with the correct environment label selector", func() {
				Expect(stanService.Spec.Selector["environment"]).To(Equal(input.Environment))
			})
			It("should create a service for stan with the correct infrastructure label selector", func() {
				Expect(stanService.Spec.Selector["infrastructure"]).To(Equal("Stan"))
			})
			It("should create a service for stan without a microservice label selector", func() {
				Expect(stanService.Spec.Selector["microservice"]).To(Equal(""))
			})
			It("should create a service for stan with the 'metrics' port exposed", func() {
				Expect(stanService.Spec.Ports[0].Name).To(Equal("metrics"))
				Expect(stanService.Spec.Ports[0].Port).To(Equal(int32(7777)))
				Expect(stanService.Spec.Ports[0].TargetPort.StrVal).To(Equal("metrics"))
			})

			// STAN StatefulSet
			It("should create a statefulset for stan named 'loismay-stan'", func() {
				object := getCreatedObject(clientSet, "StatefulSet", "loismay-stan")
				Expect(object).ToNot(BeNil())
				stanStatefulSet = object.(*appsv1.StatefulSet)
			})
			It("should create a statefulset for stan with the correct ApiVersion", func() {
				Expect(stanStatefulSet.APIVersion).To(Equal("apps/v1"))
			})
			It("should create a statefulset for stan with the correct Kind", func() {
				Expect(stanStatefulSet.Kind).To(Equal("StatefulSet"))
			})
			It("should create a statefulset for stan with the correct tenant-id annotation", func() {
				Expect(stanStatefulSet.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a statefulset for stan with the correct application-id annotation", func() {
				Expect(stanStatefulSet.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a statefulset for stan with the correct tenant label", func() {
				Expect(stanStatefulSet.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a statefulset for stan with the correct application label", func() {
				Expect(stanStatefulSet.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a statefulset for stan with the correct environment label", func() {
				Expect(stanStatefulSet.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a statefulset for stan with the correct infrastructure label", func() {
				Expect(stanStatefulSet.Labels["infrastructure"]).To(Equal("Stan"))
			})
			It("should create a statefulset for stan without a microservice label", func() {
				Expect(stanStatefulSet.Labels["microservice"]).To(Equal(""))
			})
			It("should create a statefulset for stan with the correct tenant label selector", func() {
				Expect(stanStatefulSet.Spec.Selector.MatchLabels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a statefulset for stan with the correct application label selector", func() {
				Expect(stanStatefulSet.Spec.Selector.MatchLabels["application"]).To(Equal(application.Name))
			})
			It("should create a statefulset for stan with the correct environment label selector", func() {
				Expect(stanStatefulSet.Spec.Selector.MatchLabels["environment"]).To(Equal(input.Environment))
			})
			It("should create a statefulset for stan with the correct infrastructure label selector", func() {
				Expect(stanStatefulSet.Spec.Selector.MatchLabels["infrastructure"]).To(Equal("Stan"))
			})
			It("should create a statefulset for stan without a microservice label selector", func() {
				Expect(stanStatefulSet.Spec.Selector.MatchLabels["microservice"]).To(Equal(""))
			})
			It("should create a statefulset for stan with one replica", func() {
				Expect(*stanStatefulSet.Spec.Replicas).To(Equal(int32(1)))
			})
			It("should create a pod template for stan with the correct tenant-id annotation", func() {
				Expect(stanStatefulSet.Spec.Template.Annotations["dolittle.io/tenant-id"]).To(Equal(tenant.ID))
			})
			It("should create a pod template for stan with the correct application-id annotation", func() {
				Expect(stanStatefulSet.Spec.Template.Annotations["dolittle.io/application-id"]).To(Equal(application.ID))
			})
			It("should create a pod template for stan with the correct tenant label", func() {
				Expect(stanStatefulSet.Spec.Template.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a pod template for stan with the correct application label", func() {
				Expect(stanStatefulSet.Spec.Template.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a pod template for stan with the correct environment label", func() {
				Expect(stanStatefulSet.Spec.Template.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a pod template for stan with the correct infrastructure label", func() {
				Expect(stanStatefulSet.Spec.Template.Labels["infrastructure"]).To(Equal("Stan"))
			})
			It("should create a pod template for stan without a microservice label", func() {
				Expect(stanStatefulSet.Spec.Template.Labels["microservice"]).To(Equal(""))
			})
			It("should create a pod template for stan with the stan config map as a volume", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
				Expect(stanStatefulSet.Spec.Template.Spec.Volumes[0].ConfigMap.Name).To(Equal("loismay-stan"))
			})
			It("should create a container for stan named 'stan'", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Name).To(Equal("stan"))
			})
			It("should create a container for stan with the correct image", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Image).To(Equal("nats-streaming:0.22.0"))
			})
			It("should create a container for stan with the correct arguments", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Args[0]).To(Equal("-sc"))
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Args[1]).To(Equal("/etc/stan-config/stan.conf"))
			})
			It("should create a container for stan with the 'monitor' port exposed", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(int32(8222)))
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Ports[0].Name).To(Equal("monitor"))
			})
			It("should create a container for stan with the 'metrics' port exposed", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Ports[1].ContainerPort).To(Equal(int32(7777)))
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Ports[1].Name).To(Equal("metrics"))
			})
			It("should create a container for stan with the 'POD_NAME' environmental variable set", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("POD_NAME"))
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Env[0].ValueFrom.FieldRef.FieldPath).To(Equal("metadata.name"))
			})
			It("should create a container for stan with the 'POD_NAMESPACE' environmental variable set", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Env[1].Name).To(Equal("POD_NAMESPACE"))
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].Env[1].ValueFrom.FieldRef.FieldPath).To(Equal("metadata.namespace"))
			})
			It("should create a container for stan with '/etc/stan-config' mounted", func() {
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/etc/stan-config"))
				Expect(stanStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(Equal("config-volume"))
			})

			// RawDataLogIngestor Ingress
			It("should create an ingress for rawdatalogingestor named 'ernestbush'", func() {
				object := getCreatedObject(clientSet, "Ingress", fmt.Sprintf("loismay-ernestbush-%s", customerTenantID[0:7]))
				Expect(object).ToNot(BeNil())
				rawDataLogIngestorIngress = object.(*netv1.Ingress)
			})
			It("should create an ingress for rawdatalogingestor with the production certmanager issuer", func() {
				Expect(rawDataLogIngestorIngress.Annotations["cert-manager.io/cluster-issuer"]).To(Equal("letsencrypt-staging"))
			})
			It("should create an ingress for rawdatalogingestor with the correct tenant id", func() {
				// TODO this is broken now, as we get a dynamic id
				Expect(rawDataLogIngestorIngress.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"]).To(Equal("proxy_set_header Tenant-ID \"f4679b71-1215-4a60-8483-53b0d5f2bb47\";\n"))
			})
			It("should create an ingress for rawdatalogingestor with one rule", func() {
				Expect(len(rawDataLogIngestorIngress.Spec.Rules)).To(Equal(1))
			})
			It("should create an ingress for rawdatalogingestor with the correct rule host", func() {
				Expect(rawDataLogIngestorIngress.Spec.Rules[0].Host).To(Equal("some-fancy.domain.name"))
			})
			It("should create an ingress for rawdatalogingestor with one rule path", func() {
				Expect(len(rawDataLogIngestorIngress.Spec.Rules[0].HTTP.Paths)).To(Equal(1))
			})
			It("should create an ingress for rawdatalogingestor with the correct rule path", func() {
				Expect(rawDataLogIngestorIngress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/api/not-webhooks-just-to-be-sure"))
			})
			It("should create an ingress for rawdatalogingestor with the correct rule pathtype", func() {
				Expect(string(*rawDataLogIngestorIngress.Spec.Rules[0].HTTP.Paths[0].PathType)).To(Equal("SpecialTypeNotActuallySupported"))
			})
			It("should create an ingress for rawdatalogingestor with one TLS", func() {
				Expect(len(rawDataLogIngestorIngress.Spec.TLS)).To(Equal(1))
			})
			It("should create an ingress for rawdatalogingestor with the correct TLS host", func() {
				Expect(rawDataLogIngestorIngress.Spec.TLS[0].Hosts[0]).To(Equal("some-fancy.domain.name"))
			})
			It("should create an ingress for rawdatalogingestor with the correct TLS secret name", func() {
				Expect(rawDataLogIngestorIngress.Spec.TLS[0].SecretName).To(Equal("some-fancy-certificate"))
			})

			// RawDataLogIngestor Deployment
			It("should create a deployment for rawdatalog named 'loismay-ernestbush'", func() {
				object := getCreatedObject(clientSet, "Deployment", "loismay-ernestbush")
				Expect(object).ToNot(BeNil())
				rawDataLogDeployment = object.(*appsv1.Deployment)
			})
			It("should create a deployment for raw data log with the correct Kind", func() {
				Expect(rawDataLogDeployment.Kind).To(Equal("Deployment"))
			})
			It("should create a deployment for raw data log with the correct tenant-id annotation", func() {
				Expect(rawDataLogDeployment.Annotations["dolittle.io/tenant-id"]).To(Equal(input.Dolittle.TenantID))
			})
			It("should create a deployment for raw data log with the correct application-id annotation", func() {
				Expect(rawDataLogDeployment.Annotations["dolittle.io/application-id"]).To(Equal(input.Dolittle.ApplicationID))
			})
			It("should create a deployment for raw data log with the correct microservice-id annotation", func() {
				Expect(rawDataLogDeployment.Annotations["dolittle.io/microservice-id"]).To(Equal(input.Dolittle.MicroserviceID))
			})
			It("should create a deployment for raw data log with the correct microservice-kind annotation", func() {
				Expect(rawDataLogDeployment.Annotations["dolittle.io/microservice-kind"]).To(Equal(string(input.Kind)))
			})
			It("should create a deployment for raw data log with the correct tenant label", func() {
				Expect(rawDataLogDeployment.Labels["tenant"]).To(Equal(tenant.Name))
			})
			It("should create a deployment for raw data log with the correct application label", func() {
				Expect(rawDataLogDeployment.Labels["application"]).To(Equal(application.Name))
			})
			It("should create a deployment for raw data log with the correct environment label", func() {
				Expect(rawDataLogDeployment.Labels["environment"]).To(Equal(input.Environment))
			})
			It("should create a deployment for raw data log with the correct microservice label", func() {
				Expect(rawDataLogDeployment.Labels["microservice"]).To(Equal(input.Name))
			})
		})
	})
})

func getCreatedObject(clientSet *fake.Clientset, kind, name string) runtime.Object {
	for _, create := range getCreateActions(clientSet) {
		object := create.GetObject()
		if object.GetObjectKind().GroupVersionKind().Kind == kind {
			switch resource := object.(type) {
			case *corev1.ConfigMap:
				if resource.GetName() == name {
					return resource
				}
			case *corev1.Service:
				if resource.GetName() == name {
					return resource
				}
			case *appsv1.StatefulSet:
				if resource.GetName() == name {
					return resource
				}
			case *netv1.Ingress:
				if resource.GetName() == name {
					return resource
				}
			case *appsv1.Deployment:
				if resource.GetName() == name {
					return resource
				}
			}
		}
	}
	return nil
}

func getCreateActions(clientSet *fake.Clientset) []testing.CreateAction {
	var actions []testing.CreateAction
	for _, action := range clientSet.Actions() {
		if create, ok := action.(testing.CreateAction); ok {
			actions = append(actions, create)
		}
	}
	return actions
}
