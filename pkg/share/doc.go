package share

import "github.com/dolittle-entropy/platform-api/pkg/dolittle"

type DowloadLogsRepo interface {
	GetAll() []dolittle.CustomerTenant
	FindCustomerTenant(name string) (dolittle.CustomerTenant, error)
	FindApplication(tenant string, application string, environment string) (dolittle.Application, error)
	FindIngress(domain string) (dolittle.Application, error)
}

type DownloadLogsService struct {
	repo DowloadLogsRepo
}

type HttpTenant struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type HttpApplication struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	ID          string `json:"id"`
}

type HTTPDownloadLogsInput struct {
	TenantID    string `json:"tenant_id"`
	Tenant      string `json:"tenant"`
	Application string `json:"application"`
	Environment string `json:"environment"`
	FilePath    string `json:"file_path"`
}

type HTTPDownloadLogsApplicationsResponse struct {
	Tenant       HttpTenant        `json:"tenant"`
	Applications []HttpApplication `json:"applications"`
}

type HTTPDownloadLogsLatestResponse struct {
	Tenant      HttpTenant `json:"tenant"`
	Application string     `json:"application"`
	Files       []string   `json:"files"`
}

type HTTPDownloadLogsLinkResponse struct {
	Tenant      string `json:"tenant"`
	Application string `json:"application"`
	Url         string `json:"url"`
	Expires     string `json:"expire"`
}

type HTTPCustomer struct {
	Tenant       dolittle.Tenant `json:"tenant"`
	Applications []string        `json:"applications"`
	Domains      []string        `json:"domains"`
}

type HTTPDownloadLogsCustomersResponse struct {
	Customers []HTTPCustomer `json:"customers"`
}
