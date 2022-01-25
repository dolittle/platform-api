package cicd

import (
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

func (s *service) GetDevops(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars["name"] = "devops"
	mux.SetURLVars(r, vars)
	s.getServiceAccountCredentials(w, r)
}

func (s *service) getServiceAccountCredentials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	serviceAccountName := vars["name"]

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":             "GetServiceAccountCredentials",
		"credentials":        "serviceAccount",
		"customerID":         customerID,
		"applicationID":      applicationID,
		"userID":             userID,
		"serviceAccountName": serviceAccountName,
	})
	logContext.Info("requested credentials")

	serviceAccount, err := s.k8sDolittleRepo.GetServiceAccount(logContext, applicationID, serviceAccountName)
	if err != nil {
		if err == platform.ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Service account %s not found in application %s", serviceAccountName, applicationID))
			return
		}

		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	secret, err := s.k8sDolittleRepo.GetSecret(logContext, applicationID, serviceAccount.Secrets[0].Name)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, secret.Data)
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
		if err == platform.ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Secret %s not found in application %s", secretName, applicationID))
			return
		}

		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, secret.Data)
}
