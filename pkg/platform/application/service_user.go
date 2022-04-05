package application

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *Service) UserList(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")
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
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	logContext = logContext.WithField("application_id", applicationID)

	currentUsers, err := s.userAccess.GetUsers(applicationID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("get.users")

		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	users := HttpResponseAccessUsers{
		Users: make([]HttpResponseAccessUser, 0),
	}

	for _, currentUser := range currentUsers {
		users.Users = append(users.Users, HttpResponseAccessUser{
			Email: currentUser,
		})
	}

	utils.RespondWithJSON(
		w,
		http.StatusOK,
		users,
	)
}

func (s *Service) UserAdd(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")
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
		}).Error("Bad input")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	logContext.Info("Add user access to application in active directory and kratos")

	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	// Add to Kratos
	// Add to Active Azure Directory
	err = s.userAccess.AddUser(customerID, applicationID, input.Email)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("adding.user")
		// TODO do we want something better here
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
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")
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
		}).Error("Bad input")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	logContext.Info("Remove user access to application in active directory")

	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	// Add to Kratos
	// Add to Active Azure Directory
	err = s.userAccess.RemoveUser(applicationID, input.Email)
	if err != nil {
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
