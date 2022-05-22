package codegenerator

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext      logrus.FieldLogger
	k8sDolittleRepo platformK8s.K8sRepo
}

type KafkaConfigJSON struct {
	BrokerUrl string   `json:"brokerUrl"`
	Topics    []string `json:"topics"`
}

func NewService(logContext logrus.FieldLogger,
	k8sDolittleRepo platformK8s.K8sRepo) service {
	return service{
		logContext:      logContext,
		k8sDolittleRepo: k8sDolittleRepo,
	}
}

func Test() {
	accessKeyFile, _ := os.Open("/Users/gh/Desktop/secrets/accessKey.pem")
	defer accessKeyFile.Close()

	certificateFile, _ := os.Open("/Users/gh/Desktop/secrets/certificate.pem")
	defer certificateFile.Close()

	caFile, _ := os.Open("/Users/gh/Desktop/secrets/ca.pem")
	defer caFile.Close()

	c := newCodeGeneratorClient("https://localhost:7159")
	zipFileName := "foo.zip"
	kafkaConfig := KafkaConfig{
		BrokerURL:            "Foo",
		InputTopic:           "Foo",
		CommandTopic:         "Foo",
		CommandReceiptsTopic: "Foo",
		ChangeEventsTopic:    "Foo",
		AccessKey:            accessKeyFile,
		Certificate:          certificateFile,
		Ca:                   caFile,
	}
	generatedCode := c.GenerateM3ConnectorConsumer(zipFileName,
		"foo",
		"Dev",
		"foo",
		kafkaConfig)
	outputFile := fmt.Sprintf("/Users/gh/Desktop/%s", zipFileName)
	ioutil.WriteFile(outputFile, generatedCode, fs.FileMode(os.O_CREATE))
}

func (s *service) GenerateM3ConnectorConsumer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	s.logContext.WithField("Tenant-ID", customerID).
		WithField("User-ID", userID).
		WithField("ApplicationId", applicationID).
		Info("Will generatore m3connector")

	// TODO: take env as input!
	configMap, err := s.k8sDolittleRepo.GetConfigMap(applicationID, "test-kafka-files")
	if err != nil {
		s.logContext.Error(err)
		return
	}

	accessKeyFile := strings.NewReader(configMap.Data["accessKey.pem"])
	certificateFile := strings.NewReader(configMap.Data["certificate.pem"])
	caFile := strings.NewReader(configMap.Data["ca.pem"])
	cfgJson := configMap.Data["config.json"]
	var foo KafkaConfigJSON
	err = json.Unmarshal([]byte(cfgJson), &foo)
	if err != nil {
		fmt.Println(err)
	}

	/*accessKeyFile, _ := os.Open("/Users/gh/Desktop/secrets/accessKey.pem")
	defer accessKeyFile.Close()

	certificateFile, _ := os.Open("/Users/gh/Desktop/secrets/certificate.pem")
	defer certificateFile.Close()

	//caFile, _ := os.Open("/Users/gh/Desktop/secrets/ca.pem")
	//defer caFile.Close()*/

	inputTopic := "todo-add-topic"
	commandTopic := "todo-add-command-topic"
	commandReceiptsTopic := "todo-add-command-receipts-topic"
	changeEventsTopic := "todo-add-change-events-topc"
	for _, n := range foo.Topics {
		if strings.HasSuffix(n, ".input") {
			inputTopic = n
		}

		if strings.HasSuffix(n, ".commands") {
			commandTopic = n
		}

		if strings.HasSuffix(n, ".command-receipts") {
			commandReceiptsTopic = n
		}

		if strings.HasSuffix(n, ".change-events") {
			changeEventsTopic = n
		}
	}

	c := newCodeGeneratorClient("https://localhost:7159")
	zipFileName := "foo.zip"
	kafkaConfig := KafkaConfig{
		BrokerURL:            foo.BrokerUrl,
		InputTopic:           inputTopic,
		CommandTopic:         commandTopic,
		CommandReceiptsTopic: commandReceiptsTopic,
		ChangeEventsTopic:    changeEventsTopic,
		AccessKey:            accessKeyFile,
		Certificate:          certificateFile,
		Ca:                   caFile,
	}
	generatedCode := c.GenerateM3ConnectorConsumer(zipFileName,
		"foo",
		"Dev",
		"foo",
		kafkaConfig)
	s.logContext.WithField("numberOfBytes", len(generatedCode)).Info("Code is generated")

	w.Write(generatedCode)
}
