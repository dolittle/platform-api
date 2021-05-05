package microservice

import "k8s.io/client-go/kubernetes"

type Storage interface {
	Write(info HttpInputDolittle, data []byte) error
	Read(info HttpInputDolittle) ([]byte, error)
	GetAll(tenantID string, applicationID string) ([]HttpMicroserviceBase, error)
}

type service struct {
	storage    Storage
	simpleRepo simpleRepo
	k8sClient  *kubernetes.Clientset
}

const (
	Simple                 = "simple"
	BusinessMomentsAdaptor = "buisness-moments"
	Webhook                = "webhook"
)

type HttpInputMicroserviceKind struct {
	Kind string `json:"kind"`
}

type HttpMicroserviceBase struct {
	Dolittle HttpInputDolittle `json:"dolittle"`
	Name     string            `json:"name"`
	Kind     string            `json:"kind"`
	Extra    interface{}       `json:"extra"`
}
type HttpInputDolittle struct {
	ApplicationID  string `json:"applicationId"`
	TenantID       string `json:"tenantId"`
	MicroserviceID string `json:"microserviceId"`
}

type HttpInputSimpleIngress struct {
	Host             string `json:"path"`
	SecretNamePrefix string `json:"secretNamePrefix"` // Not happy with this part
	Path             string `json:"path"`
	Pathtype         string `json:"pathType"`
}

type HttpInputSimpleInfo struct {
	Dolittle HttpInputDolittle    `json:"dolittle"`
	Name     string               `json:"name"`
	Kind     string               `json:"kind"`
	Extra    HttpInputSimpleExtra `json:"extra"`
}

type HttpInputSimpleExtra struct {
	Headimage    string                 `json:"headImage"`
	Runtimeimage string                 `json:"runtimeImage"`
	Environment  string                 `json:"environment"`
	Ingress      HttpInputSimpleIngress `json:"ingress"`
}
