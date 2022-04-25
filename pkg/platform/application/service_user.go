package application

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *Service) UserList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	customerID := vars["customerID"]
	userID := r.Header.Get("User-ID")

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":      "UserList",
		"customer_id": customerID,
		"user_id":     userID,
	})

	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	logContext.Info("List Access in active directory and kratos")
	logContext = logContext.WithField("application_id", applicationID)

	application, err := s.gitRepo.GetApplication(customerID, applicationID)
	if err != nil {
		if err != storage.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("Storage has failed")
			utils.RespondWithError(w, http.StatusInternalServerError, platform.ErrStudioInfoMissing.Error())
			return
		}
		utils.RespondWithError(w, http.StatusNotFound, "Application not found for this customer")
		return
	}

	currentUsers, err := s.userAccess.GetUsers(applicationID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("get.users")

		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	response := HttpResponseAccessUsers{
		ID:    application.ID,
		Name:  application.Name,
		Users: make([]HttpResponseAccessUser, 0),
	}

	for _, currentUser := range currentUsers {
		response.Users = append(response.Users, HttpResponseAccessUser{
			Email: currentUser,
		})
	}

	utils.RespondWithJSON(
		w,
		http.StatusOK,
		response,
	)
}

func (s *Service) UserAdd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	customerID := vars["customerID"]
	userID := r.Header.Get("User-ID")

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":      "UserAdd",
		"customer_id": customerID,
		"user_id":     userID,
	})

	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	var input HttpInputAccessUser
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to read")
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to read payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Bad input")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	logContext.Info("Add user access to application in active directory and kratos")

	err = s.userAccess.AddUser(customerID, applicationID, input.Email)
	if err != nil {
		if err == user.ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, "Email not found, unable to add to application")
			return
		}

		if err == user.ErrTooManyResults {
			utils.RespondWithError(w, http.StatusUnprocessableEntity, "More than one email found, unable to add to application")
			return
		}

		if err == user.ErrEmailAlreadyExists {
			utils.RespondWithError(w, http.StatusUnprocessableEntity, "Email already has access to this application")
			return
		}

		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("adding.user")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to add user")
		return
	}

	utils.RespondWithJSON(
		w,
		http.StatusOK,
		utils.HTTPMessageResponse{
			Message: "User added",
		},
	)
}

func (s *Service) UserRemove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	customerID := vars["customerID"]
	userID := r.Header.Get("User-ID")

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":      "UserAdd",
		"customer_id": customerID,
		"user_id":     userID,
	})

	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	var input HttpInputAccessUser
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to read")
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to read payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Bad input")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	logContext.Info("Remove user access to application in active directory")

	err = s.userAccess.RemoveUser(applicationID, input.Email)
	if err != nil {
		if err == user.ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, "Email not found, unable to remove from application")
			return
		}

		if err == user.ErrTooManyResults {
			utils.RespondWithError(w, http.StatusUnprocessableEntity, "More than one email found, unable to remove from application")
			return
		}

		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("remove.user")
		// TODO do we want something better here
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to remove user")
		return
	}

	// Remove from Active Azure Directory
	utils.RespondWithJSON(
		w,
		http.StatusOK,
		utils.HTTPMessageResponse{
			Message: "User removed",
		},
	)
}
