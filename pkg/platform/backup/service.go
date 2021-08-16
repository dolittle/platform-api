package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	azureHelpers "github.com/dolittle-entropy/platform-api/pkg/azure"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewService(k8sDolittleRepo platform.K8sRepo, k8sClient *kubernetes.Clientset) service {
	return service{
		k8sDolittleRepo: k8sDolittleRepo,
	}
}

func (s *service) GetLatestByApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "GetLatestByApplication",
		"tenantID":      tenantID,
		"applicationID": applicationID,
		"environment":   environment,
	})

	tenantInfo, err := s.gitRepo.GetTerraformTenant(tenantID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "s.gitRepo.GetTerraformTenant(tenantID)",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	applicationInfo, err := s.gitRepo.GetApplication(tenantID, applicationID)
	if err == nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "s.gitRepo.GetApplication(tenantID, applicationID)",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusBadRequest, "Application already exists")
		return
	}

	exists := funk.Contains(applicationInfo.Environments, func(item platform.HttpInputEnvironment) bool {
		found := false
		if item.Name == environment {
			found = true
		}
		return found
	})

	if exists {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Environment %s already exists", environment))
		return
	}

	namespace := fmt.Sprintf("application-%s", applicationID)

	azureStorageInfo, err := getStorageAccountInfo(r.Context(), namespace, s.k8sClient)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "getStorageAccountInfo",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	azureShareName, err := getShareName(ctx, namespace, s.k8sClient, metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("tenant=%s,application=%s,environment=%s,infrastructure=Mongo", tenantInfo.Name, applicationInfo.Name, environment),
	})

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "getShareName",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	latest, err := azureHelpers.LatestX(azureStorageInfo.Name, azureStorageInfo.Key, azureShareName)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "azureHelpers.LatestX(azureStorageInfo.Name, azureStorageInfo.Key, azureShareName)",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, HTTPDownloadLogsLatestResponse{
		Tenant: HttpTenant{
			Name: tenantInfo.Name,
			ID:   tenantInfo.GUID,
		},
		Application: applicationInfo.Name,
		Files:       latest.Files,
	})
}

func (s *service) CreateLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	var input HTTPDownloadLogsInput
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "CreateLink",
		"tenantID":      input.TenantID,
		"applicationID": input.Application,
		"environment":   input.Environment,
	})

	tenantInfo, err := s.gitRepo.GetTerraformTenant(input.TenantID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "s.gitRepo.GetTerraformTenant(tenantID)",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	applicationInfo, err := s.gitRepo.GetApplication(input.TenantID, input.Application)
	if err == nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "s.gitRepo.GetApplication(tenantID, applicationID)",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusBadRequest, "Application already exists")
		return
	}

	exists := funk.Contains(applicationInfo.Environments, func(item platform.HttpInputEnvironment) bool {
		found := false
		if item.Name == input.Environment {
			found = true
		}
		return found
	})

	if exists {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Environment %s already exists", input.Environment))
		return
	}

	// Create link
	namespace := fmt.Sprintf("application-%s", applicationID)
	azureStorageInfo, err := getStorageAccountInfo(r.Context(), namespace, s.k8sClient)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "getStorageAccountInfo",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	azureShareName, err := getShareName(ctx, namespace, s.k8sClient, metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("tenant=%s,application=%s,environment=%s,infrastructure=Mongo", tenantInfo.Name, applicationInfo.Name, environment),
	})

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "getShareName",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	checkShareName := fmt.Sprintf("/%s/mongo", azureShareName)
	if !strings.HasPrefix(input.FilePath, checkShareName) {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, "Not valid for this application")
		return
	}

	// Cleanup path
	filePath := input.FilePath
	filePath = strings.TrimLeft(filePath, "/")
	filePath = strings.TrimLeft(filePath, azureShareName)
	filePath = strings.TrimLeft(filePath, "/")

	expires := time.Now().UTC().Add(48 * time.Hour)

	url, err := azureHelpers.CreateLink(azureStorageInfo.Name, azureStorageInfo.Key, azureShareName, filePath, expires)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "azureHelpers.CreateLink",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, HTTPDownloadLogsLinkResponse{
		Tenant:      tenantInfo.Name,
		Application: applicationInfo.Name,
		Url:         url,
		Expires:     expires.Format(time.RFC3339Nano),
	})
}

func getStorageAccountInfo(ctx context.Context, namespace string, client *kubernetes.Clientset) (AzureStorageInfo, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, "storage-account-secret", metaV1.GetOptions{})
	if err != nil {
		return AzureStorageInfo{}, err
	}
	return AzureStorageInfo{
		Name: string(secret.Data["azurestorageaccountname"]),
		Key:  string(secret.Data["azurestorageaccountkey"]),
	}, nil
}

func getShareName(ctx context.Context, namespace string, client *kubernetes.Clientset, opts metaV1.ListOptions) (string, error) {
	crons, err := client.BatchV1beta1().CronJobs(namespace).List(ctx, opts)
	if err != nil {
		return "", err
	}

	if len(crons.Items) == 0 {
		return "not-found", nil
	}
	return crons.Items[0].Spec.JobTemplate.Spec.Template.Spec.Volumes[0].AzureFile.ShareName, nil
}
