package cicd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext      logrus.FieldLogger
	k8sDolittleRepo platform.K8sRepo
}

func NewService(logContext logrus.FieldLogger, k8sDolittleRepo platform.K8sRepo) *service {
	s := &service{
		logContext:      logContext,
		k8sDolittleRepo: k8sDolittleRepo,
	}
	return s
}

func (s *service) GetServiceAccountCredentials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "GetServiceAccountCredentials",
		"credentials":   "serviceAccount",
		"customerID":    customerID,
		"applicationID": applicationID,
		"userID":        userID,
	})
	logContext.Info("requested credentials")

	secretName := "azure-devops"
	secret, err := s.k8sDolittleRepo.GetSecret(logContext, applicationID, secretName)
	if err != nil {
		if err == platform.SecretNotFound {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Secret %s not found in application %s", secretName, applicationID))
			return
		}

		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	b, _ := json.Marshal(secret.Data)
	utils.RespondWithJSON(w, http.StatusOK, string(b))
}

func (s *service) GetContainerRegistryCredentials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "GetServiceAccountCredentials",
		"credentials":   "containerRegistry",
		"customerID":    customerID,
		"applicationID": applicationID,
		"userID":        userID,
	})
	logContext.Info("requested credentials")

	secretName := "acr"
	secret, err := s.k8sDolittleRepo.GetSecret(logContext, applicationID, secretName)
	if err != nil {
		if err == platform.SecretNotFound {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Secret %s not found in application %s", secretName, applicationID))
			return
		}

		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	b, _ := json.Marshal(secret.Data)
	utils.RespondWithJSON(w, http.StatusOK, string(b))
}
