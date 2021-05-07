package microservice

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"k8s.io/client-go/kubernetes"
)

type Storage interface {
	Write(info HttpInputDolittle, data []byte) error
	Read(info HttpInputDolittle) ([]byte, error)
	GetAll(tenantID string, applicationID string) ([]HttpMicroserviceBase, error)
}

type service struct {
	storage                    Storage
	simpleRepo                 simpleRepo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	k8sDolittleRepo            platform.K8sRepo
	k8sClient                  *kubernetes.Clientset
}

const (
	Simple                 = "simple"
	BusinessMomentsAdaptor = "buisness-moments-adaptor"
	Webhook                = "webhook"
)

type HttpInputMicroserviceKind struct {
	Kind string `json:"kind"`
}

type HttpMicroserviceBase struct {
	Dolittle    HttpInputDolittle `json:"dolittle"`
	Name        string            `json:"name"`
	Kind        string            `json:"kind"`
	Environment string            `json:"environment"`
	Extra       interface{}       `json:"extra"`
}
type HttpInputDolittle struct {
	ApplicationID  string `json:"applicationId"`
	TenantID       string `json:"tenantId"`
	MicroserviceID string `json:"microserviceId"`
}

type HttpInputSimpleIngress struct {
	Host             string `json:"host"`
	SecretNamePrefix string `json:"secretNamePrefix"` // Not happy with this part
	Path             string `json:"path"`
	Pathtype         string `json:"pathType"`
}

type HttpInputSimpleInfo struct {
	Dolittle    HttpInputDolittle    `json:"dolittle"`
	Name        string               `json:"name"`
	Kind        string               `json:"kind"`
	Environment string               `json:"environment"`
	Extra       HttpInputSimpleExtra `json:"extra"`
}

type HttpInputSimpleExtra struct {
	Headimage    string                 `json:"headImage"`
	Runtimeimage string                 `json:"runtimeImage"`
	Ingress      HttpInputSimpleIngress `json:"ingress"`
}

type HttpInputBusinessMomentAdaptorInfo struct {
	Dolittle    HttpInputDolittle                   `json:"dolittle"`
	Name        string                              `json:"name"`
	Kind        string                              `json:"kind"`
	Environment string                              `json:"environment"`
	Extra       HttpInputBusinessMomentAdaptorExtra `json:"extra"`
}

type HttpInputBusinessMomentAdaptorExtra struct {
	Headimage    string                 `json:"headImage"`
	Runtimeimage string                 `json:"runtimeImage"`
	Ingress      HttpInputSimpleIngress `json:"ingress"`
	Connector    interface{}            `json:"connector"`
}

type HttpInputBusinessMomentAdaptorConnectorWebhook struct {
	Kind   string                                               `json:"kind"`
	Config HttpInputBusinessMomentAdaptorConnectorWebhookConfig `json:"config"`
}

type HttpInputBusinessMomentAdaptorConnectorWebhookConfig struct {
	Kind   string      `json:"kind"`
	Config interface{} `json:"config"`
}

type HttpInputBusinessMomentAdaptorConnectorWebhookConfigBasic struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HttpInputBusinessMomentAdaptorConnectorWebhookConfigBearer struct {
	Token string `json:"token"`
}

type HttpResponseMicroservices struct {
	Application   platform.ShortInfo                  `json:"application"`
	Microservices []platform.ShortInfoWithEnvironment `json:"microservices"`
}

type HttpResponsePodStatus struct {
	Application  platform.ShortInfo                  `json:"application"`
	Microservice platform.ShortInfoWithEnvironment   `json:"microservice"`
	PodStatus    []platform.ShortInfoWithEnvironment `json:"podStatus"`
}

type HttpResponsePodLog struct {
	ApplicationID  string `json:"applicationId"`
	MicroserviceID string `json:"microserviceId"`
	PodName        string `json:"podName"`
	Logs           string `json:"logs"`
}
