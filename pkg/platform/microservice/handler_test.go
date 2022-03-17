package microservice_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/storage"
)

var _ = Describe("Handler", func() {
	When("creating microservices", func() {
		Context("that are simple", func() {
			XIt("should return a 200 when it succeeds", func() {
				var (
					mockStorageRepo storage.Repo
					k8sDolittleRepo platformK8s.K8sRepo
					k8sClient       kubernetes.Interface
					simpleRepo      simple.Repo
					logContext      logrus.FieldLogger
					tenantID        string
				)

				handler := microservice.NewHandler(false, mockStorageRepo, k8sDolittleRepo, k8sClient, simpleRepo, logContext)
				microservice := platform.HttpMicroserviceBase{
					MicroserviceBase: platform.MicroserviceBase{
						Dolittle: platform.HttpInputDolittle{
							ApplicationID:  "test",
							CustomerID:     "test",
							MicroserviceID: "test",
						},
						Name:        "test",
						Kind:        platform.MicroserviceKindSimple,
						Environment: "test",
					},
				}
				tenantID = "test"

				recorder := httptest.NewRecorder()
				httpHandler := http.HandlerFunc(handler.Create)
				requestJson, _ := json.Marshal(microservice)
				request := httptest.NewRequest("POST", "", bytes.NewBuffer(requestJson))
				request.Header.Add("Tenant-ID", tenantID)

				httpHandler.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
		})
	})
})
