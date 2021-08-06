package share

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	azureHelpers "github.com/dolittle-entropy/platform-api/pkg/azure"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
)

func NewLogsService(repo DowloadLogsRepo) DownloadLogsService {
	return DownloadLogsService{
		repo: repo,
	}
}

func (s *DownloadLogsService) GetCustomers(w http.ResponseWriter, r *http.Request) {
	response := HTTPDownloadLogsCustomersResponse{}
	customers := s.repo.GetAll()
	for _, customer := range customers {
		applications := []string{}
		domains := []string{}
		for _, application := range customer.Applications {
			applications = append(applications, application.Name)
			domains = append(domains, application.IngressHosts...)
		}

		response.Customers = append(response.Customers, HTTPCustomer{
			Tenant:       customer.Tenant,
			Applications: applications,
			Domains:      domains,
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *DownloadLogsService) GetApplicationsByTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenant := vars["tenant"]

	customer, err := s.repo.FindCustomerTenant(tenant)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	applications := make([]HttpApplication, 0)
	for _, application := range customer.Applications {
		applications = append(applications, HttpApplication{
			ID:          application.ID,
			Name:        application.Name,
			Environment: application.Environment,
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, HTTPDownloadLogsApplicationsResponse{
		Tenant: HttpTenant{
			Name: customer.Tenant.Name,
			ID:   customer.Tenant.ID,
		},
		Applications: applications,
	})
}

func (s *DownloadLogsService) GetLatestByApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// TODO this is not clear
	environment := vars["environment"]
	applicationName := vars["application"]
	tenant := vars["tenant"]
	application, err := s.repo.FindApplication(tenant, applicationName, environment)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Application not found")
		return
	}

	// TODO find domain by applicationName would also need env
	customer, err := s.repo.FindCustomerTenant(application.Tenant.Name)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	if customer.AzureStorageAccount.Name == "" {
		utils.RespondWithError(w, http.StatusNotFound, "Azure not in use")
		return
	}

	if customer.AzureStorageAccount.Key == "XXX" {
		utils.RespondWithError(w, http.StatusInternalServerError, "not supporting customer.AzureStorageAccount.Key, XXX")
		return
	}

	// TODO add context?
	latest, err := azureHelpers.LatestX(customer.AzureStorageAccount.Name, customer.AzureStorageAccount.Key, application.AzureShareName)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, HTTPDownloadLogsLatestResponse{
		Tenant: HttpTenant{
			Name: customer.Tenant.Name,
			ID:   customer.Tenant.ID,
		},
		Application: application.Name,
		Files:       latest.Files,
	})
}

func (s *DownloadLogsService) GetLatestByDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domain := vars["domain"]
	application, err := s.repo.FindIngress(domain)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Domain not found")
		return
	}

	customer, err := s.repo.FindCustomerTenant(application.Tenant.Name)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	if customer.AzureStorageAccount.Name == "" {
		utils.RespondWithError(w, http.StatusNotFound, "Azure not in use")
		return
	}

	if customer.AzureStorageAccount.Key == "XXX" {
		utils.RespondWithError(w, http.StatusInternalServerError, "not supporting customer.AzureStorageAccount.Key, XXX")
		return
	}

	// TODO add context?
	latest, err := azureHelpers.LatestX(customer.AzureStorageAccount.Name, customer.AzureStorageAccount.Key, application.AzureShareName)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, HTTPDownloadLogsLatestResponse{
		Tenant: HttpTenant{
			Name: customer.Tenant.Name,
			ID:   customer.Tenant.ID,
		},
		Application: application.Name,
		Files:       latest.Files,
	})
}

func (s *DownloadLogsService) CreateLink(w http.ResponseWriter, r *http.Request) {
	var input HTTPDownloadLogsInput
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	customer, err := s.repo.FindCustomerTenant(input.TenantID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	application, err := s.repo.FindApplication(customer.Tenant.ID, input.Application, input.Environment)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Tenant with application not found")
		return
	}
	// Create link

	checkShareName := fmt.Sprintf("/%s/mongo", application.AzureShareName)
	if !strings.HasPrefix(input.FilePath, checkShareName) {
		fmt.Println(input)
		fmt.Println(checkShareName)
		utils.RespondWithError(w, http.StatusUnprocessableEntity, "Not valid for this application")
		return
	}

	// Cleanup path
	filePath := input.FilePath
	filePath = strings.TrimLeft(filePath, "/")
	filePath = strings.TrimLeft(filePath, application.AzureShareName)
	filePath = strings.TrimLeft(filePath, "/")

	expires := time.Now().UTC().Add(48 * time.Hour)

	url, err := azureHelpers.CreateLink(customer.AzureStorageAccount.Name, customer.AzureStorageAccount.Key, application.AzureShareName, filePath, expires)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, HTTPDownloadLogsLinkResponse{
		Tenant:      customer.Tenant.Name,
		Application: application.Name,
		Url:         url,
		Expires:     expires.Format(time.RFC3339Nano),
	})
}
