package microservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/requesthandler"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	authV1 "k8s.io/api/authorization/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func NewService(gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo, k8sClient kubernetes.Interface, logContext logrus.FieldLogger) service {
	parser := requesthandler.NewJsonParser()

	return service{
		gitRepo:         gitRepo,
		k8sDolittleRepo: k8sDolittleRepo,
		parser:          parser,
		handlers:        requesthandler.CreateHandlers(parser, k8sClient, gitRepo, k8sDolittleRepo, logContext),
	}
}

// TODO https://dolittle.freshdesk.com/a/tickets/1352 how to add multiple entries to ingress
func (s *service) Create(responseWriter http.ResponseWriter, request *http.Request) {
	var microserviceBase platform.HttpMicroserviceBase
	requestBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = json.Unmarshal(requestBytes, &microserviceBase)

	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer request.Body.Close()

	userID := request.Header.Get("User-ID")
	applicationID := microserviceBase.Dolittle.ApplicationID
	environment := microserviceBase.Environment

	applicationInfo, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(responseWriter, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithJSON(responseWriter, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Application %s not found", applicationID),
		})
		return
	}

	customer := k8s.Tenant{
		ID:   applicationInfo.Tenant.ID,
		Name: applicationInfo.Tenant.Name,
	}

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(responseWriter, customer.ID, applicationID, userID)
	if !allowed {
		return
	}

	if !s.gitRepo.IsAutomationEnabled(customer.ID, applicationID, environment) {
		utils.RespondWithError(
			responseWriter,
			http.StatusBadRequest,
			fmt.Sprintf(
				"Customer %s with application %s in environment %s does not allow changes via Studio",
				customer.ID,
				applicationID,
				environment,
			),
		)
		return
	}

	handler, err := s.handlers.GetForKind(microserviceBase.Kind)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
	}
	ms, createError := handler.Create(request, requestBytes, applicationInfo)
	if createError != nil {
		utils.RespondWithError(responseWriter, createError.StatusCode, createError.Error())
	}
	utils.RespondWithJSON(responseWriter, http.StatusOK, ms)

}

func (s *service) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	tenantInfo, err := s.gitRepo.GetTerraformTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenant := k8s.Tenant{
		ID:   tenantInfo.GUID,
		Name: tenantInfo.Name,
	}

	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]

	data, err := s.gitRepo.GetMicroservice(
		tenant.ID,
		applicationID,
		environment,
		microserviceID,
	)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response interface{}
	json.Unmarshal(data, &response)
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetByApplicationID(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	tenantInfo, err := s.gitRepo.GetTerraformTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenant := k8s.Tenant{
		ID:   tenantInfo.GUID,
		Name: tenantInfo.Name,
	}

	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	data, err := s.gitRepo.GetMicroservices(tenant.ID, applicationID)

	if err != nil {
		// TODO change
		utils.RespondWithJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, data)
}

func (s *service) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// I feel we shouldn't need namespace
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	if !s.gitRepo.IsAutomationEnabled(tenantID, applicationID, environment) {
		utils.RespondWithError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf(
				"Tenant %s with application %s in environment %s does not allow changes via Studio",
				tenantID,
				applicationID,
				environment,
			),
		)
		return
	}

	errStr := ""
	statusCode := http.StatusOK

	// Hairy stuff
	msData, err := s.gitRepo.GetMicroservice(tenantID, applicationID, environment, microserviceID)

	if err == nil {
		// LOG THIS
		var whatKind platform.HttpInputMicroserviceKind
		err = json.Unmarshal(msData, &whatKind)
		if err == nil {
			handler, err := s.handlers.GetForKind(whatKind.Kind)
			if err == nil {
				err = handler.Delete(namespace, microserviceID)
			}

			if err != nil {
				statusCode = http.StatusUnprocessableEntity
				errStr = err.Error()
			}
		}
	}

	if statusCode == http.StatusOK {
		err = s.gitRepo.DeleteMicroservice(tenantID, applicationID, environment, microserviceID)
		if err != nil {
			statusCode = http.StatusUnprocessableEntity
			errStr = err.Error()
		}
	}

	utils.RespondWithJSON(w, statusCode, map[string]string{
		"namespace":       namespace,
		"error":           errStr,
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"action":          "Remove microservice",
	})
}

func (s *service) GetLiveByApplicationID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	application, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	microservices, err := s.k8sDolittleRepo.GetMicroservices(applicationID)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := platform.HttpResponseMicroservices{
		Application: platform.ShortInfo{
			Name: application.Name,
			ID:   application.ID,
		},
		Microservices: microservices,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetPodStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := strings.ToLower(vars["environment"])

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	status, err := s.k8sDolittleRepo.GetPodStatus(applicationID, microserviceID, environment)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//response := HttpResponsePodStatus{
	//	Application: platform.ShortInfo{
	//		Name: application.Name,
	//		ID:   application.ID,
	//	},
	//	Microservice: platform.ShortInfoWithEnvironment{
	//		Name: application.Name,
	//		ID:   application.ID,
	//	},
	//}

	utils.RespondWithJSON(w, http.StatusOK, status)
}

func (s *service) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	podName := vars["podName"]
	containerName := r.FormValue("containerName")
	// TODO how to ignore?
	if containerName == "" {
		containerName = "head"
	}

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	logData, err := s.k8sDolittleRepo.GetLogs(applicationID, containerName, podName)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, platform.HttpResponsePodLog{
		ApplicationID:  applicationID,
		MicroserviceID: "TODO",
		PodName:        podName,
		Logs:           logData,
	})
}

func (s *service) GetConfigMap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	configMapName := vars["configMapName"]

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	//contentType := strings.ToLower(r.Header.Get("Content-Type"))

	download := r.FormValue("download")
	fileType := r.FormValue("fileType")
	if fileType == "" {
		fileType = "json"
	}

	filterFileType := funk.ContainsString([]string{
		"json",
		"yaml",
	}, fileType)

	if !filterFileType {
		utils.RespondWithError(w, http.StatusBadRequest, "File-Type not supported")
		return
	}

	// Hmm this will let them see things they are not allowed to see.
	// But it wont let them update it
	// TODO when / if we allow update, we will need more protection
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	configMap, err := s.k8sDolittleRepo.GetConfigMap(applicationID, configMapName)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Config map %s not found in application %s", configMapName, applicationID))
		return
	}

	// Little hack to work around https://github.com/kubernetes/client-go/issues/861
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"

	output, _ := yaml.Marshal(configMap)

	if download == "1" {
		contentType := ""
		switch fileType {
		case "json":
			contentType = "application/json"

		case "yaml":
			contentType = "application/yaml"
		}

		fileName := fmt.Sprintf("%s.%s", configMapName, fileType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%s`, fileName))
		w.Header().Set("Content-Type", contentType)
	}

	switch fileType {
	case "json":
		utils.RespondWithJSON(w, http.StatusOK, configMap)
	case "yaml":
		utils.RespondWithYAML(w, http.StatusOK, output)
	}

}

func (s *service) GetSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	secretName := vars["secretName"]

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")

	download := r.FormValue("download")
	fileType := r.FormValue("fileType")

	// TODO today being very strict about we allow
	poormansSecretCheck := funk.Contains([]string{
		"-secret-env-variables",
	}, func(filter string) bool {
		return strings.HasSuffix(secretName, filter)
	})

	if !poormansSecretCheck {
		utils.RespondWithError(w, http.StatusForbidden, "Secret endpoint is currently locked down to secrets supported by Dolittle microservices")
		return
	}

	if fileType == "" {
		fileType = "json"
	}

	filterFileType := funk.ContainsString([]string{
		"json",
		"yaml",
	}, fileType)

	if !filterFileType {
		utils.RespondWithError(w, http.StatusBadRequest, "File-Type not supported")
		return
	}

	// Hmm this will let them see things they are not allowed to see.
	// But it wont let them update it
	// TODO when / if we allow update, we will need more protection
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	secret, err := s.k8sDolittleRepo.GetSecret(applicationID, secretName)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Config map %s not found in application %s", secret, applicationID))
		return
	}

	// Little hack to work around https://github.com/kubernetes/client-go/issues/861
	secret.APIVersion = "v1"
	secret.Kind = "Secret"

	output, _ := yaml.Marshal(secret)

	if download == "1" {
		contentType := ""
		switch fileType {
		case "json":
			contentType = "application/json"

		case "yaml":
			contentType = "application/yaml"
		}

		fileName := fmt.Sprintf("%s.%s", secret, fileType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%s`, fileName))
		w.Header().Set("Content-Type", contentType)
	}

	switch fileType {
	case "json":
		utils.RespondWithJSON(w, http.StatusOK, secret)
	case "yaml":
		utils.RespondWithYAML(w, http.StatusOK, output)
	}

}

func (s *service) CanI(w http.ResponseWriter, r *http.Request) {
	// To get fine control, we need to lookup the AD group which can be added via terrform
	// TODO we do not create rbac roles today? meaning this look up will break.
	// TODO the can access, wont work with new customers due to the rbac setup
	// TODO bringing online the ad group from microsoft will allow us to check group access
	type httpInput struct {
		UserID            string `json:"user_id"`
		TenantID          string `json:"tenant_id"`
		Group             string `json:"group"`
		ApplicationID     string `json:"application_id"`
		ResourceAttribute struct {
			Verb     string `json:"verb"`
			Resource string `json:"resource"`
			Name     string `json:"name"`
		} `json:"resourceAttribute"`
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var input httpInput
	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	attribute := authV1.ResourceAttributes{
		Namespace: fmt.Sprintf("application-%s", input.ApplicationID),
		Verb:      "list",
		Resource:  "pods",
	}

	if input.ResourceAttribute.Resource != "" {
		attribute.Resource = input.ResourceAttribute.Resource
	}

	if input.ResourceAttribute.Verb != "" {
		attribute.Verb = input.ResourceAttribute.Verb
	}

	allowed, err := s.k8sDolittleRepo.CanModifyApplicationWithResourceAttributes(input.TenantID, input.ApplicationID, input.UserID, attribute)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allowedStr := "allowed"
	if !allowed {
		allowedStr = "denied"
	}
	utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{
		"allowed": allowedStr,
	})
}
