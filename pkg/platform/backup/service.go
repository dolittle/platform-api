package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	azureHelpers "github.com/dolittle/platform-api/pkg/azure"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewService(logContext logrus.FieldLogger, gitRepo storage.Repo, k8sDolittleRepo platformK8s.K8sRepo, k8sClient kubernetes.Interface) service {
	return service{
		logContext:      logContext,
		gitRepo:         gitRepo,
		k8sDolittleRepo: k8sDolittleRepo,
		k8sClient:       k8sClient,
	}
}

func (s *service) GetLatestByApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	// Not made to lower due to how we query kubernetes (for now)
	environment := vars["environment"]

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "GetLatestByApplication",
		"tenantID":      tenantID,
		"applicationID": applicationID,
		"environment":   environment,
	})

	applicationInfo, err := s.gitRepo.GetApplication(tenantID, applicationID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "s.gitRepo.GetApplication(tenantID, applicationID)",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusBadRequest, "Application already exists")
		return
	}

	exists := environmentExists(applicationInfo.Environments, environment)

	if !exists {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Environment %s does not exist", environment))
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
		LabelSelector: fmt.Sprintf("environment=%s,infrastructure=Mongo", environment),
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
		Application: platform.ShortInfo{
			ID:   applicationInfo.ID,
			Name: applicationInfo.Name,
		},
		Environment: environment,
		Files:       latest.Files,
	})
}

func (s *service) CreateLink(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	ctx := r.Context()

	var input HTTPDownloadLogsInput
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "CreateLink",
		"tenantID":      tenantID,
		"applicationID": input.ApplicationID,
		"environment":   input.Environment,
	})

	applicationInfo, err := s.gitRepo.GetApplication(tenantID, input.ApplicationID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"where": "s.gitRepo.GetApplication(tenantID, applicationID)",
		}).Error("lookup error")
		utils.RespondWithError(w, http.StatusBadRequest, "Application already exists")
		return
	}

	exists := environmentExists(applicationInfo.Environments, input.Environment)

	if !exists {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Environment %s does not exist", input.Environment))
		return
	}

	// Create link
	namespace := fmt.Sprintf("application-%s", input.ApplicationID)
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
		LabelSelector: fmt.Sprintf("environment=%s,infrastructure=Mongo", input.Environment),
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
		Application: platform.ShortInfo{
			ID:   applicationInfo.ID,
			Name: applicationInfo.Name,
		},
		Url:     url,
		Expires: expires.Format(time.RFC3339Nano),
	})
}

func environmentExists(environments []platform.HttpInputEnvironment, environment string) bool {
	return funk.Contains(environments, func(item platform.HttpInputEnvironment) bool {
		return item.Name == environment
	})
}

// TODO maybe we should move this into the k8sRepo
func getStorageAccountInfo(ctx context.Context, namespace string, client kubernetes.Interface) (AzureStorageInfo, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, "storage-account-secret", metaV1.GetOptions{})
	if err != nil {
		return AzureStorageInfo{}, err
	}
	return AzureStorageInfo{
		Name: string(secret.Data["azurestorageaccountname"]),
		Key:  string(secret.Data["azurestorageaccountkey"]),
	}, nil
}

func getShareName(ctx context.Context, namespace string, client kubernetes.Interface, opts metaV1.ListOptions) (string, error) {
	crons, err := client.BatchV1beta1().CronJobs(namespace).List(ctx, opts)
	if err != nil {
		return "", err
	}

	if len(crons.Items) == 0 {
		return "", errors.New("not-found")
	}
	return crons.Items[0].Spec.JobTemplate.Spec.Template.Spec.Volumes[0].AzureFile.ShareName, nil
}
