package platform

type HttpInputApplication struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenantId"`
}

type HttpInputEnvironment struct {
	Name              string `json:"name"`
	DomainPrefix      string `json:"domainPrefix"`
	Host              string `json:"host"`
	TenantID          string `json:"tenantId"`
	ApplicationID     string `json:"applicationId"`
	AutomationEnabled bool   `json:"automationEnabled"`
}

type HttpResponseApplication2 struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	TenantID      string                 `json:"tenantId"`
	Environments  []HttpInputEnvironment `json:"environments"`
	Microservices []HttpMicroserviceBase `json:"microservices"`
}

type HttpResponseApplication struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	TenantID     string                 `json:"tenantId"`
	Environments []HttpInputEnvironment `json:"environments"`
}

type HttpResponseApplications struct {
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	Applications []ShortInfoWithEnvironment `json:"applications"`
}

type ImageInfo struct {
	Image string `json:"image"`
	Name  string `json:"name"`
}

type MicroserviceInfo struct {
	Name        string      `json:"name"`
	Environment string      `json:"environment"`
	ID          string      `json:"id"`
	Images      []ImageInfo `json:"images"`
}
type PodInfo struct {
	Name       string      `json:"name"`
	Phase      string      `json:"phase"`
	Containers []ImageInfo `json:"containers"`
}

type PodData struct {
	Namespace    string    `json:"namespace"`
	Microservice ShortInfo `json:"microservice"`
	Pods         []PodInfo `json:"pods"`
}

type Tenant struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Ingress struct {
	Host        string `json:"host"`
	Environment string `json:"environment"`
	Path        string `json:"path"`
}

type Application struct {
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	Tenant    Tenant    `json:"tenant"`
	Ingresses []Ingress `json:"ingresses"`
}

type ShortInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ShortInfoWithEnvironment struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	ID          string `json:"id"`
}

type GitRepo interface {
	Write(tenantID string, applicationID string, data []byte) error
	Read(tenantID string, applicationID string) ([]byte, error)
	GetAll(tenantID string) ([]Application, error)
}

const (
	Simple                 = "simple"
	BusinessMomentsAdaptor = "business-moments-adaptor"
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
	Moments      []BusinessMoment       `json:"moments"`
	Entities     []Entity               `json:"entities"`
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
	Application   ShortInfo          `json:"application"`
	Microservices []MicroserviceInfo `json:"microservices"`
}

type HttpResponsePodStatus struct {
	Application  ShortInfo                  `json:"application"`
	Microservice ShortInfoWithEnvironment   `json:"microservice"`
	PodStatus    []ShortInfoWithEnvironment `json:"podStatus"`
}

type HttpResponsePodLog struct {
	ApplicationID  string `json:"applicationId"`
	MicroserviceID string `json:"microserviceId"`
	PodName        string `json:"podName"`
	Logs           string `json:"logs"`
}

type TerraformCustomer struct {
	Name                    string `json:"name"`
	GUID                    string `json:"guid"`
	AzureStorageAccountName string `json:"azure_storage_account_name"`
	ContainerRegistryName   string `json:"container_registry_name"`
}

type StudioConfig struct {
	BuildOverwrite         bool     `json:"build_overwrite"`
	AutomationEnabled      bool     `json:"automation_enabled"` // Crude beginning of knowing if the customer allows  automation
	AutomationEnvironments []string `json:"automation_environments"`
}

type Entity struct {
	Name          string `json:"name"`
	TypeID        string `json:"typeId"` // Should this be eventType, GUUID
	IdForRetrival string `json:"idForRetrival"`
	Filter        string `json:"filter"`
	Transform     string `json:"transform"`
}

type BusinessMoment struct {
	EntityID  string `json:"entityId"`
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
	Filter    string `json:"filter"`
	Transform string `json:"transform"`
}

type HttpInputEntity struct {
	ApplicationID  string `json:"applicationId"`
	Environment    string `json:"environment"`
	MicroserviceID string `json:"microserviceId"`
	Entity         Entity `json:"entity"`
}

type HttpInputBusinessMoment struct {
	ApplicationID  string         `json:"applicationId"`
	Environment    string         `json:"environment"`
	MicroserviceID string         `json:"microserviceId"`
	Moment         BusinessMoment `json:"moment"`
}

type HttpResponseBusinessMoments struct {
	ApplicationID string `json:"application_id"`
	Environment   string `json:"environment"`
	//MicroserviceID string                    `json:"microservice_id"` // Could omit if empty
	Moments  []HttpInputBusinessMoment `json:"moments"`
	Entities []HttpInputEntity         `json:"entities"`
}