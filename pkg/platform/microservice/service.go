package microservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/parser"
	"github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle/platform-api/pkg/platform/microservice/rawdatalog"
	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	authv1 "k8s.io/api/authorization/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func NewService(
	isProduction bool,
	gitRepo storage.Repo,
	k8sDolittleRepo platformK8s.K8sRepo,
	k8sClient kubernetes.Interface,
	logContext logrus.FieldLogger,
) service {
	parser := parser.NewJsonParser()
	rawDataLogRepo := rawdatalog.NewRawDataLogIngestorRepo(isProduction, k8sDolittleRepo, k8sClient, logContext)
	specFactory := purchaseorderapi.NewK8sResourceSpecFactory()
	k8sResources := purchaseorderapi.NewK8sResource(k8sClient, specFactory)

	return service{
		gitRepo:                    gitRepo,
		simpleRepo:                 k8sSimple.NewSimpleRepo(k8sClient, k8sDolittleRepo, isProduction),
		businessMomentsAdaptorRepo: NewBusinessMomentsAdaptorRepo(k8sClient, isProduction),
		rawDataLogIngestorRepo:     rawDataLogRepo,
		k8sDolittleRepo:            k8sDolittleRepo,
		parser:                     parser,
		purchaseOrderHandler: purchaseorderapi.NewHandler(
			parser,
			purchaseorderapi.NewRepo(k8sResources, specFactory, k8sClient),
			gitRepo,
			rawDataLogRepo,
			logContext),
		logContext: logContext,
	}
}

// TODO https://dolittle.freshdesk.com/a/tickets/1352 how to add multiple entries to ingress
func (s *service) Create(w http.ResponseWriter, request *http.Request) {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method": "Create",
	})
	// Parse JSON
	requestBytes, microserviceBase, err := s.readMicroserviceBase(request, logContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Errorf("Invalid request payload: %w", err).Error())
		return
	}
	defer request.Body.Close()

	// Confirm customer exists
	customerID := request.Header.Get("Tenant-ID")
	applicationID := microserviceBase.Dolittle.ApplicationID
	studioInfo, err := storage.GetStudioInfo(s.gitRepo, customerID, applicationID, logContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	customer := k8s.Tenant{
		ID:   studioInfo.TerraformCustomer.GUID,
		Name: studioInfo.TerraformCustomer.Name,
	}

	// Confirm application exists
	storedApplication, err := s.gitRepo.GetApplication(customer.ID, applicationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Not able to find application in the storage")
		return
	}

	// TODO Confirm the application exists
	// storedApplication.Status.State == storage.BuildStatusStateFinishedSuccess
	// - is it in terraform?
	// - is it being made

	// Confirm user has access to this application + customer
	userID := request.Header.Get("User-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customer.ID, applicationID, userID)
	if !allowed {
		return
	}

	environment := microserviceBase.Environment
	exists := storage.EnvironmentExists(storedApplication.Environments, environment)

	if !exists {
		utils.RespondWithError(w, http.StatusBadRequest, "Unable to add a microservice to an environment that does not exist")
		return
	}

	// TODO is this legacy?
	// Check if automation enabled
	if !s.gitRepo.IsAutomationEnabledWithStudioConfig(studioInfo.StudioConfig, applicationID, environment) {
		utils.RespondWithError(
			w,
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

	customerTenants := make([]platform.CustomerTenantInfo, 0)
	for _, envInfo := range storedApplication.Environments {
		if envInfo.Name != environment {
			continue
		}

		customerTenants = envInfo.CustomerTenants
		break
	}

	// TODO add paths
	// Refactor this
	// TODO do we need to confirm the application is in the name space?
	// Should this really come from the cluster?
	applicationInfo, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Application %s not found", applicationID),
		})
		return
	}

	// TODO check path
	switch microserviceBase.Kind {
	case platform.MicroserviceKindSimple:
		s.handleSimpleMicroservice(w, request, requestBytes, applicationInfo, customerTenants)
	// TODO let us get simple to work and we can come back to these others.
	//case platform.MicroserviceKindBusinessMomentsAdaptor:
	//	s.handleBusinessMomentsAdaptor(w, request, requestBytes, applicationInfo, customerTenants)
	//case platform.MicroserviceKindRawDataLogIngestor:
	//s.handleRawDataLogIngestor(w, request, requestBytes, applicationInfo, customerTenants)
	//case platform.MicroserviceKindPurchaseOrderAPI:
	//	purchaseOrderAPI, err := s.purchaseOrderHandler.Create(requestBytes, applicationInfo)
	//	if err != nil {
	//		utils.RespondWithError(w, err.StatusCode, err.Error())
	//		break
	//	}
	//	utils.RespondWithJSON(w, http.StatusAccepted, purchaseOrderAPI)
	default:
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Kind %s is not supported", microserviceBase.Kind))
	}
}

func (s *service) Update(w http.ResponseWriter, request *http.Request) {
	customerID := request.Header.Get("Tenant-ID")

	logContext := s.logContext.WithFields(logrus.Fields{
		"method": "Update",
	})
	requestBytes, microserviceBase, err := s.readMicroserviceBase(request, logContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Errorf("Invalid request payload: %w", err).Error())
		return
	}
	defer request.Body.Close()

	applicationID := microserviceBase.Dolittle.ApplicationID
	environment := microserviceBase.Environment

	studioInfo, err := storage.GetStudioInfo(s.gitRepo, customerID, applicationID, logContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	applicationInfo, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Application %s not found", applicationID),
		})
		return
	}

	customer := k8s.Tenant{
		ID:   studioInfo.TerraformCustomer.GUID,
		Name: studioInfo.TerraformCustomer.Name,
	}

	userID := request.Header.Get("User-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customer.ID, applicationID, userID)
	if !allowed {
		return
	}

	if !s.gitRepo.IsAutomationEnabledWithStudioConfig(studioInfo.StudioConfig, applicationID, environment) {
		utils.RespondWithError(
			w,
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

	switch microserviceBase.Kind {
	case platform.MicroserviceKindPurchaseOrderAPI:
		// TODO handle other updation operations too
		purchaseOrderAPI, err := s.purchaseOrderHandler.UpdateWebhooks(requestBytes, applicationInfo)
		if err != nil {
			utils.RespondWithError(w, err.StatusCode, err.Error())
		}
		utils.RespondWithJSON(w, http.StatusOK, purchaseOrderAPI)
	default:
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Kind %s is not supported", microserviceBase.Kind))
	}
}

func (s *service) readMicroserviceBase(request *http.Request, logContext *logrus.Entry) (requestBytes []byte, microserviceBase platform.HttpMicroserviceBase, err error) {
	microserviceBase = platform.HttpMicroserviceBase{}
	requestBytes, err = ioutil.ReadAll(request.Body)
	if err != nil {
		logContext.WithError(err).Error("Failed to read request body")
		err = fmt.Errorf("failed to read request body: %w", err)
		return
	}

	if err = json.Unmarshal(requestBytes, &microserviceBase); err != nil {
		logContext.WithError(err).Error("Failed to read json from request body")
		err = fmt.Errorf("failed to read json from request body: %w", err)
		return
	}
	return
}

func (s *service) GetByID(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")
	tenantInfo, err := s.gitRepo.GetTerraformTenant(customerID)
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
	customerID := r.Header.Get("Tenant-ID")
	tenantInfo, err := s.gitRepo.GetTerraformTenant(customerID)
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
	logContext := s.logContext.WithFields(logrus.Fields{
		"method": "Delete",
	})
	vars := mux.Vars(r)
	// I feel we shouldn't need namespace
	applicationID := vars["applicationID"]
	// TODO Can we rely on this? or does it have to be exact?
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]
	namespace := fmt.Sprintf("application-%s", applicationID)

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	studioInfo, err := storage.GetStudioInfo(s.gitRepo, customerID, applicationID, logContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	if !s.gitRepo.IsAutomationEnabledWithStudioConfig(studioInfo.StudioConfig, applicationID, environment) {
		utils.RespondWithError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf(
				"Tenant %s with application %s in environment %s does not allow changes via Studio",
				customerID,
				applicationID,
				environment,
			),
		)
		return
	}

	errStr := ""
	statusCode := http.StatusOK

	// Hairy stuff
	msData, err := s.gitRepo.GetMicroservice(customerID, applicationID, environment, microserviceID)

	if err == nil {
		// LOG THIS
		var whatKind platform.HttpInputMicroserviceKind
		err = json.Unmarshal(msData, &whatKind)
		if err == nil {
			switch whatKind.Kind {
			case platform.MicroserviceKindSimple:
				err = s.simpleRepo.Delete(applicationID, environment, microserviceID)
			case platform.MicroserviceKindBusinessMomentsAdaptor:
				err = s.businessMomentsAdaptorRepo.Delete(applicationID, environment, microserviceID)
			case platform.MicroserviceKindRawDataLogIngestor:
				// TODO add environment
				err = s.rawDataLogIngestorRepo.Delete(namespace, microserviceID)
			case platform.MicroserviceKindPurchaseOrderAPI:
				err = s.purchaseOrderHandler.Delete(applicationID, environment, microserviceID)
			}
			if err != nil {
				statusCode = http.StatusUnprocessableEntity
				errStr = err.Error()
			}
		}
	}

	if statusCode == http.StatusOK {
		err = s.gitRepo.DeleteMicroservice(customerID, applicationID, environment, microserviceID)
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
	customerID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
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
	customerID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	status, err := s.k8sDolittleRepo.GetPodStatus(applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

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
	customerID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
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
	customerID := r.Header.Get("Tenant-ID")
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
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
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
	customerID := r.Header.Get("Tenant-ID")

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
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":     "GetSecret",
		"userID":     userID,
		"customerID": customerID,
	})

	secret, err := s.k8sDolittleRepo.GetSecret(logContext, applicationID, secretName)
	if err != nil {
		if err == platformK8s.ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Secret %s not found in application %s", secretName, applicationID))
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
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

	attribute := authv1.ResourceAttributes{
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

func (s *service) Restart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := strings.ToLower(vars["environment"])

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	err := s.k8sDolittleRepo.RestartMicroservice(applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Microservice restarted",
	})
}
