package staticfiles

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type service struct {
	logContext      logrus.FieldLogger
	k8sClient       kubernetes.Interface
	k8sDolittleRepo platform.K8sRepo
}

func NewService(k8sDolittleRepo platform.K8sRepo, k8sClient kubernetes.Interface, logContext logrus.FieldLogger) service {
	return service{
		logContext:      logContext,
		k8sDolittleRepo: k8sDolittleRepo,
		k8sClient:       k8sClient,
	}
}

func (s *service) GetAll(responseWriter http.ResponseWriter, request *http.Request) {
	tenantID := request.Header.Get("Tenant-ID")

	vars := mux.Vars(request)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]

	logContext := s.logContext.WithFields(logrus.Fields{
		"service":        "staticFiles",
		"method":         "GetAll",
		"tenantID":       tenantID,
		"applicationID":  applicationID,
		"environment":    environment,
		"microserviceID": microserviceID,
	})

	logContext.Info("Get all files")

	upstream, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, "Failed to lookup microservice dns")
		return
	}

	sharedSecret, _ := s.getSharedSecret(applicationID, environment, microserviceID)
	request.URL.Path = "/manage/list-files"
	request.RequestURI = "/manage/list-files"

	request.Header = http.Header{}
	request.Header.Set("x-shared-secret", sharedSecret)

	ctx, cancel := context.WithTimeout(request.Context(), 10*time.Second)
	defer cancel()
	request = request.WithContext(ctx)
	serveReverseProxy(upstream, responseWriter, request)
}

func (s *service) Add(responseWriter http.ResponseWriter, request *http.Request) {
	tenantID := request.Header.Get("Tenant-ID")

	vars := mux.Vars(request)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]

	logContext := s.logContext.WithFields(logrus.Fields{
		"service":        "staticFiles",
		"method":         "Add",
		"tenantID":       tenantID,
		"applicationID":  applicationID,
		"environment":    environment,
		"microserviceID": microserviceID,
	})

	logContext.Info("Add file")

	parts := strings.Split(
		request.URL.Path,
		fmt.Sprintf("staticFiles/%s/add/", microserviceID),
	)
	// TODO check if filename is set
	fileName := parts[1]

	upstream, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, "Failed to lookup microservice dns")
		return
	}

	sharedSecret, _ := s.getSharedSecret(applicationID, environment, microserviceID)
	request.URL.Path = fmt.Sprintf("/manage/add/%s", fileName)
	request.RequestURI = fmt.Sprintf("/manage/add/%s", fileName)

	request.Header = http.Header{}
	request.Header.Set("x-shared-secret", sharedSecret)

	// TODO maybe we want longer on upload
	ctx, cancel := context.WithTimeout(request.Context(), 10*time.Second)
	defer cancel()
	request = request.WithContext(ctx)
	serveReverseProxy(upstream, responseWriter, request)
}

func (s *service) Remove(responseWriter http.ResponseWriter, request *http.Request) {
	tenantID := request.Header.Get("Tenant-ID")

	vars := mux.Vars(request)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]

	logContext := s.logContext.WithFields(logrus.Fields{
		"service":        "staticFiles",
		"method":         "Add",
		"tenantID":       tenantID,
		"applicationID":  applicationID,
		"environment":    environment,
		"microserviceID": microserviceID,
	})

	logContext.Info("Remove file")

	parts := strings.Split(
		request.URL.Path,
		fmt.Sprintf("staticFiles/%s/remove/", microserviceID),
	)
	// TODO check if filename is set
	fileName := parts[1]

	upstream, err := s.k8sDolittleRepo.GetMicroserviceDNS(applicationID, microserviceID)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, "Failed to lookup microservice dns")
		return
	}

	sharedSecret, _ := s.getSharedSecret(applicationID, environment, microserviceID)
	request.URL.Path = fmt.Sprintf("/manage/remove/%s", fileName)
	request.RequestURI = fmt.Sprintf("/manage/remove/%s", fileName)

	request.Header = http.Header{}
	request.Header.Set("x-shared-secret", sharedSecret)

	// TODO maybe we want longer on upload
	ctx, cancel := context.WithTimeout(request.Context(), 10*time.Second)
	defer cancel()
	request = request.WithContext(ctx)
	serveReverseProxy(upstream, responseWriter, request)
}

func serveReverseProxy(host string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse("/")
	url.Host = host
	// Hard coding to http for now
	url.Scheme = "http"
	proxy := httputil.NewSingleHostReverseProxy(url)
	fmt.Println(req)
	// Update the request
	req.Host = url.Host
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme

	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

func (s *service) getSharedSecret(applicationID string, environment string, microserviceID string) (string, error) {
	client := s.k8sClient
	ctx := context.TODO()
	namespace := fmt.Sprintf("application-%s", applicationID)
	secrets, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	for _, secret := range secrets.Items {
		annotations := secret.GetAnnotations()

		if annotations["dolittle.io/microservice-kind"] != string(platform.MicroserviceKindStaticFilesV1) {
			continue
		}

		if annotations["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		if !strings.HasSuffix(secret.ObjectMeta.Name, "-secret-env-variables") {
			continue
		}

		return string(secret.Data["HEADER_SECRET"]), nil
	}

	if err != nil {
		return "", err
	}
	return "", nil
}
