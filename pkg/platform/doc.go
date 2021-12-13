package platform

import (
	"errors"

	networkingv1 "k8s.io/api/networking/v1"
)

type Microservice interface {
	GetBase() MicroserviceBase
}
type MicroserviceBase struct {
	Dolittle    HttpInputDolittle `json:"dolittle"`
	Name        string            `json:"name"`
	Kind        MicroserviceKind  `json:"kind"`
	Environment string            `json:"environment"`
}

func (m MicroserviceBase) GetBase() MicroserviceBase {
	return m
}

type HttpInputApplication struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenantId"`
}

type TenantId string

type EnvironmentIngress struct {
	Host         string `json:"host"`
	DomainPrefix string `json:"domainPrefix"`
	SecretName   string `json:"secretName"`
}

type EnvironmentIngresses map[TenantId]EnvironmentIngress

type HttpInputEnvironment struct {
	AutomationEnabled bool                 `json:"automationEnabled"`
	Name              string               `json:"name"`
	TenantID          string               `json:"tenantId"`
	ApplicationID     string               `json:"applicationId"`
	Tenants           []TenantId           `json:"tenants"`
	Ingresses         EnvironmentIngresses `json:"ingresses"`
}

type HttpResponseApplication struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	TenantID      string                 `json:"tenantId"`
	TenantName    string                 `json:"tenantName"`
	Environments  []HttpInputEnvironment `json:"environments"`
	Microservices []HttpMicroserviceBase `json:"microservices,omitempty"`
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

type ContainerStatusInfo struct {
	Image    string `json:"image"`
	Name     string `json:"name"`
	Age      string `json:"age"`
	State    string `json:"state"`
	Started  string `json:"started"`
	Restarts int32  `json:"restarts"`
}

type MicroserviceInfo struct {
	Name         string                           `json:"name"`
	Environment  string                           `json:"environment"`
	ID           string                           `json:"id"`
	Images       []ImageInfo                      `json:"images"`
	Kind         string                           `json:"kind"`
	IngressURLS  []IngressURLWithCustomerTenantID `json:"ingressUrls"`
	IngressPaths []networkingv1.HTTPIngressPath   `json:"ingressPaths"`
}
type PodInfo struct {
	Name       string                `json:"name"`
	Phase      string                `json:"phase"`
	Containers []ContainerStatusInfo `json:"containers"`
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

type MicroserviceKind string

const (
	MicroserviceKindSimple                 MicroserviceKind = "simple"
	MicroserviceKindBusinessMomentsAdaptor MicroserviceKind = "business-moments-adaptor"
	MicroserviceKindRawDataLogIngestor     MicroserviceKind = "raw-data-log-ingestor"
	MicroserviceKindPurchaseOrderAPI       MicroserviceKind = "purchase-order-api" // TODO purchase-order-api VS purchase-order
)

type HttpInputMicroserviceKind struct {
	Kind MicroserviceKind `json:"kind"`
}

type HttpMicroserviceBase struct {
	MicroserviceBase
	Extra interface{} `json:"extra"`
}
type HttpInputDolittle struct {
	ApplicationID  string `json:"applicationId"`
	TenantID       string `json:"tenantId"`
	MicroserviceID string `json:"microserviceId"`
}

type HttpInputSimpleIngress struct {
	Host             string `json:"host"`
	SecretNamePrefix string `json:"secretNamePrefix"` // Not happy with this part
	DomainPrefix     string `json:"domainPrefix"`     // Not happy with this part
	Path             string `json:"path"`
	Pathtype         string `json:"pathType"`
}

type HttpInputSimpleInfo struct {
	MicroserviceBase
	Extra HttpInputSimpleExtra `json:"extra"`
}

type HttpInputSimpleExtra struct {
	Headimage    string                 `json:"headImage"`
	Runtimeimage string                 `json:"runtimeImage"`
	Ingress      HttpInputSimpleIngress `json:"ingress"`
}

type HttpInputBusinessMomentAdaptorInfo struct {
	MicroserviceBase
	Extra HttpInputBusinessMomentAdaptorExtra `json:"extra"`
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

type HttpInputRawDataLogIngestorInfo struct {
	MicroserviceBase
	Extra HttpInputRawDataLogIngestorExtra `json:"extra"`
}

type HttpInputRawDataLogIngestorExtra struct {
	Headimage                 string                             `json:"headImage"`
	Runtimeimage              string                             `json:"runtimeImage"`
	Ingress                   HttpInputRawDataLogIngestorIngress `json:"ingress"`
	Webhooks                  []RawDataLogIngestorWebhookConfig  `json:"webhooks"`
	WebhookStatsAuthorization string                             `json:"webhookStatsAuthorization"`
	WriteTo                   string                             `json:"writeTo"`
}

type HttpInputRawDataLogIngestorIngress struct {
	Host         string `json:"host"`
	DomainPrefix string `json:"domainPrefix"`
	Path         string `json:"path"`
	Pathtype     string `json:"pathType"`
}

type RawDataLogIngestorWebhookConfig struct {
	Kind          string `json:"kind"`
	UriSuffix     string `json:"uriSuffix"`
	Authorization string `json:"authorization"`
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
	Name                      string `json:"name"`
	GUID                      string `json:"guid"`
	AzureStorageAccountName   string `json:"azure_storage_account_name"`
	AzureStorageAccountKey    string `json:"azure_storage_account_key"`
	ContainerRegistryName     string `json:"container_registry_name"`
	ContainerRegistryPassword string `json:"container_registry_password"`
	ContainerRegistryUsername string `json:"container_registry_username"`
}

type StudioConfig struct {
	BuildOverwrite       bool     `json:"build_overwrite"`
	DisabledEnvironments []string `json:"disabled_environments"`
}

type Entity struct {
	Name              string `json:"name"`
	EntityTypeID      string `json:"entityTypeId"`
	IdNameForRetrival string `json:"idNameForRetrival"`
	FilterCode        string `json:"filterCode"`
	TransformCode     string `json:"transformCode"`
}

type BusinessMoment struct {
	EntityTypeID   string `json:"entityTypeId"`
	Name           string `json:"name"`
	UUID           string `json:"uuid"`
	EmbeddingCode  string `json:"embeddingCode"`
	ProjectionCode string `json:"projectionCode"`
}

type HttpInputBusinessMomentEntity struct {
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
	Moments  []HttpInputBusinessMoment       `json:"moments"`
	Entities []HttpInputBusinessMomentEntity `json:"entities"`
}

type HttpInputPurchaseOrderInfo struct {
	MicroserviceBase
	Extra HttpInputPurchaseOrderExtra `json:"extra"`
}

type HttpInputPurchaseOrderExtra struct {
	Headimage      string                             `json:"headImage"`
	Runtimeimage   string                             `json:"runtimeImage"`
	Ingress        HttpInputRawDataLogIngestorIngress `json:"ingress"`
	Webhooks       []RawDataLogIngestorWebhookConfig  `json:"webhooks"`
	RawDataLogName string                             `json:"rawDataLogName"`
}

var (
	ErrNotFound = errors.New("not-found")
)

type TerraformApplication struct {
	Customer struct {
		GUID string `json:"guid"`
		Name string `json:"name"`
	} `json:"customer"`
	GroupID string `json:"group_id"`
	GUID    string `json:"guid"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
}

/*
- db.getCollection("stream-processor-states").find({IsFailing: { $exists: false }})
- Get eventstores eventstore_01
- Get stream-processor-states
- Failing streams
	db.getCollection("stream-processor-states").find({IsFailing: { $exists: false }, FailingPartitions: {$ne: {}}})
- State
	db.getCollection("stream-processor-states").find({})
- Event log count
	db.getCollection("event-log").count()
*/
type RuntimeStreamStates struct {
	Event       string `json:"applicationId"`
	Environment string `json:"environment"`
}
type HttpResponseRuntimeStreamStates struct {
	ApplicationID  string                `json:"applicationId"`
	Environment    string                `json:"environment"`
	MicroserviceID string                `json:"microserviceId"`
	Data           []RuntimeStreamStates `json:"data"`
}

type RuntimeLatestEvent struct {
	Row         string `bson:"row,omitempty"`
	EventTypeId string `bson:"eventTypeId,omitempty"`
	Occurred    string `bson:"occurred,omitempty"`
}

type RuntimeState struct {
	Position                  string                         `json:"position"`
	EventProcessor            string                         `json:"eventProcessor"`
	SourceStream              string                         `json:"sourceStream"`
	FailingPartitions         []RuntimeStateFailingPartition `json:"failingPartitions,omitempty"`
	LastSuccessfullyProcessed string                         `json:"lastSuccessfullyProcessed"`
	Kind                      string                         `json:"kind"`
}

type RuntimeStateFailingPartition struct {
	LastFailed         string `json:"lastFailed"`
	Partition          string `json:"partition"`
	Position           string `json:"position"`
	ProcessingAttempts int32  `json:"processingAttempts"`
	Reason             string `json:"reason"`
	RetryTime          string `json:"retryTime"`
}

type PurchaseOrderStatus struct {
	Status              string `json:"status"`
	LastReceivedPayload string `json:"lastReceivedPayload"`
	Error               string `json:"error"`
}

type IngressURLWithCustomerTenantID struct {
	URL              string `json:"url"`
	CustomerTenantID string `json:"customerTenantID"`
}
