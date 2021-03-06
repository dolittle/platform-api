package platform

import (
	"errors"

	authv1 "k8s.io/api/authorization/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

const (
	TodoCustomersTenantID        string = "17426336-fb8e-4425-8ab7-07d488367be9"
	DevelopmentCustomersTenantID string = "445f8ea8-1a6f-40d7-b2fc-796dba92dc44"
)

var (
	ErrStudioInfoMissing = errors.New("studio info is missing, reach out to the platform team")
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

type HttpResponsePersonalisedInfo struct {
	ResourceGroup               string                                `json:"resourceGroup"`
	ClusterName                 string                                `json:"clusterName"`
	SubscriptionID              string                                `json:"subscriptionId"`
	ApplicationID               string                                `json:"applicationId"`
	ContainerRegistryName       string                                `json:"containerRegistryName"`
	Endpoints                   HttpResponsePersonalisedInfoEndpoints `json:"endpoints"`
	RawSubjectRulesReviewStatus authv1.SubjectRulesReviewStatus       `json:"subjectRulesReviewStatus"`
}

type HttpResponseKubernetesResourceAccess struct {
	Secrets    []string `json:"secrets"`
	ConfigMaps []string `json:"configMaps"`
}

type HttpResponsePersonalisedInfoEndpoints struct {
	Cluster           string `json:"cluster"`
	ContainerRegistry string `json:"containerRegistry"`
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
	Host             string `json:"host"`
	Environment      string `json:"environment"`
	Path             string `json:"path"`
	CustomerTenantID string `json:"customerTenantID"`
}

type Application struct {
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	Customer  Tenant    `json:"customer"`
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

type MicroserviceKind string

const (
	MicroserviceKindSimple                 MicroserviceKind = "simple"
	MicroserviceKindUnknown                MicroserviceKind = "unknown"
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
	CustomerID     string `json:"customerId"`
	MicroserviceID string `json:"microserviceId"`
}

type HttpInputSimpleIngress struct {
	Path     string `json:"path"`
	Pathtype string `json:"pathType"`
}

type HttpInputSimpleInfo struct {
	MicroserviceBase
	Extra HttpInputSimpleExtra `json:"extra"`
}

type HttpInputSimpleExtra struct {
	Headimage    string                     `json:"headImage"`
	HeadPort     int32                      `json:"headPort"`
	Runtimeimage string                     `json:"runtimeImage"`
	Ingress      HttpInputSimpleIngress     `json:"ingress"`
	Ispublic     bool                       `json:"isPublic"`
	Headcommand  HttpInputSimpleCommand     `json:"headCommand"`
	Connections  HttpEnvironmentConnections `json:"connections"`
}

type HttpInputSimpleCommand struct {
	Command []string `json:"command"`
	Args    []string `json:"args"`
}

type HttpEnvironmentConnections struct {
	M3Connector bool `json:"m3Connector"`
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
	Headimage                 string                            `json:"headImage"`
	Runtimeimage              string                            `json:"runtimeImage"`
	Ingress                   HttpInputSimpleIngress            `json:"ingress"`
	Webhooks                  []RawDataLogIngestorWebhookConfig `json:"webhooks"`
	WebhookStatsAuthorization string                            `json:"webhookStatsAuthorization"`
	WriteTo                   string                            `json:"writeTo"`
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
	GUID                      string `json:"guid"` // This is customerID
	AzureStorageAccountName   string `json:"azure_storage_account_name"`
	AzureStorageAccountKey    string `json:"azure_storage_account_key"`
	ContainerRegistryName     string `json:"container_registry_name"`
	ContainerRegistryPassword string `json:"container_registry_password"`
	ContainerRegistryUsername string `json:"container_registry_username"`
	ModuleName                string `json:"module_name"`
	PlatformEnvironment       string `json:"platform_environment"`
}

type StudioConfig struct {
	BuildOverwrite       bool     `json:"build_overwrite"`
	DisabledEnvironments []string `json:"disabled_environments"`
	CanCreateApplication bool     `json:"can_create_application"`
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
	Headimage      string                            `json:"headImage"`
	Runtimeimage   string                            `json:"runtimeImage"`
	Ingress        HttpInputSimpleIngress            `json:"ingress"`
	Webhooks       []RawDataLogIngestorWebhookConfig `json:"webhooks"`
	RawDataLogName string                            `json:"rawDataLogName"`
}

type TerraformApplication struct {
	// TODO change from this and use CustomerID
	Customer      TerraformCustomer `json:"customer"`
	GroupID       string            `json:"group_id"`
	GUID          string            `json:"guid"` // This is applicationID
	ApplicationID string            `json:"application_id"`
	Kind          string            `json:"kind"`
	Name          string            `json:"name"`
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

type StudioEnvironmentVariable struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

type HttpResponseEnvironmentVariables struct {
	ApplicationID  string                      `json:"applicationId"`
	MicroserviceID string                      `json:"microserviceId"`
	Environment    string                      `json:"environment"`
	Data           []StudioEnvironmentVariable `json:"data"`
}
type HttpResponseConfigFilesNamesList struct {
	ApplicationID  string   `json:"applicationId"`
	MicroserviceID string   `json:"microserviceId"`
	Environment    string   `json:"environment"`
	Data           []string `json:"data"`
}

type HttpResponseDeleteConfigFile struct {
	ApplicationID  string `json:"applicationId"`
	MicroserviceID string `json:"microserviceId"`
	Environment    string `json:"environment"`
	Success        bool   `json:"success"`
}

type HttpRequestDeleteConfigFile struct {
	ApplicationID  string `json:"applicationId"`
	MicroserviceID string `json:"microserviceId"`
	Environment    string `json:"environment"`
	Key            string `json:"key"`
}

type MicroserviceMetadataShortInfo struct {
	CustomerID       string `json:"customerId"`
	CustomerName     string `json:"customerName"`
	ApplicationID    string `json:"applicationId"`
	ApplicationName  string `json:"applicationName"`
	Environment      string `json:"environment"`
	MicroserviceID   string `json:"microserviceId"`
	MicroserviceName string `json:"microserviceName"`
}

type RuntimeTenantsIDS map[string]interface{}

// CustomerTenantInfo
// Today this is specific for high level customer tenant info,
// This does not contain all the low level data across all our customers,
// We rely on code for that, at the moment
type CustomerTenantInfo struct {
	Alias            string               `json:"alias"`
	Environment      string               `json:"environment"`
	CustomerTenantID string               `json:"customerTenantId"`
	Hosts            []CustomerTenantHost `json:"hosts"`

	MicroservicesRel []CustomerTenantMicroserviceRel `json:"microservicesRel"`
}

const (
	HostNotInSystem = "na"
	HostLookUp      = ""
)

type CustomerTenantHost struct {
	Host string `json:"host"`
	// If empty get it from the cluster?
	// If not in the cluster make
	// na = not in the cluster
	SecretName string `json:"secretName"`
}

type CustomerTenantMicroserviceRel struct {
	MicroserviceID string `json:"microserviceId"`
	// ffb20e4f_a74fed4a
	// (microserviceID first block + customerTenantID first block)
	Hash string `json:"hash"`
	// We could have a legacy option here to include the current data, but why?
}

type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
