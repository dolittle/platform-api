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

	upstreamURL := fmt.Sprintf("%s/manage/list-files", upstream)
	request.URL.Path = upstreamURL

	request.Header.Del("Tenant-ID")
	request.Header.Del("User-ID")
	request.Header.Set("x-shared-secret", sharedSecret)

	// TODO change the url
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request = request.WithContext(ctx)
	serveReverseProxy(upstream, responseWriter, request)
}

func (s *service) Add(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")

	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]

	//logger := s.logger.WithFields(logrus.Fields{
	//	"service":        "PurchaseOrderAPI",
	//	"method":         "GetDataStatus",
	//	"tenantID":       tenantID,
	//	"applicationID":  applicationID,
	//	"environment":    environment,
	//	"microserviceID": microserviceID,
	//})

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"tenant_id":       tenantID,
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
	})
}

func serveReverseProxy(host string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse("/")
	url.Host = host
	// Hard coding to http for now
	url.Scheme = "http"
	proxy := httputil.NewSingleHostReverseProxy(url)

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

		// secret-en v-variables
		//if annotations["dolittle.io/microservice-kind"] != plat
		// the microserviceID is unique per microservice so that's enough for the check
		if annotations["dolittle.io/microservice-kind"] != string(platform.MicroserviceKindStaticFilesV1) {
			continue
		}

		if annotations["dolittle.io/microservice-id"] != microserviceID {
			continue
		}
		fmt.Println(secret)
	}

	if err != nil {
		return "", err
	}
	return "", nil
}
