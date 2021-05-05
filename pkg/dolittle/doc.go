package dolittle

type AzureStorageInfo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}
type Tenant struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Application struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Tenant         Tenant         `json:"tenant"`
	Environment    string         `json:"environment"`
	Microservices  []Microservice `json:"microservices"`
	TenantIDS      []string       `json:"tenant_ids"`
	AzureShareName string         `json:"azure_share_name"`
	IngressHosts   []string       `json:"ingress_hosts"`
}

type Microservice struct {
	Name        string `json:"name"`
	Application string `json:"application"`
	Environment string `json:"environment"`
}

type CustomerTenant struct {
	Namespace           string           `json:"namespace"`
	ApplicationID       string           `json:"application_id"`
	Tenant              Tenant           `json:"tenant"`
	AzureStorageAccount AzureStorageInfo `json:"azure_storage_account"`

	Applications []Application `json:"applications"`
}
