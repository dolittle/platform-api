package codegenerator

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func (s *service) GenerateM3ConnectorConsumer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := vars["environment"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	s.logContext.WithField("Tenant-ID", customerID).
		WithField("User-ID", userID).
		WithField("ApplicationId", applicationID).
		Info("Will generatore m3connector")

	configMap, err := s.k8sDolittleRepo.GetConfigMap(applicationID, fmt.Sprintf("%s-kafka-files", environment))
	if err != nil {
		s.logContext.Error(err)
		return
	}

	accessKeyFile := strings.NewReader(configMap.Data["accessKey.pem"])
	certificateFile := strings.NewReader(configMap.Data["certificate.pem"])
	caFile := strings.NewReader(configMap.Data["ca.pem"])
	cfgJson := configMap.Data["config.json"]
	var kafkaTopicsCfg KafkaConfigJSON
	err = json.Unmarshal([]byte(cfgJson), &kafkaTopicsCfg)
	if err != nil {
		fmt.Println(err)
	}

	inputTopic := "todo-add-topic"
	commandTopic := "todo-add-command-topic"
	commandReceiptsTopic := "todo-add-command-receipts-topic"
	changeEventsTopic := "todo-add-change-events-topc"
	for _, n := range kafkaTopicsCfg.Topics {
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
	zipFileName := "m3connector-consumer.zip"
	kafkaConfig := KafkaConfig{
		BrokerURL:            kafkaTopicsCfg.BrokerUrl,
		InputTopic:           inputTopic,
		CommandTopic:         commandTopic,
		CommandReceiptsTopic: commandReceiptsTopic,
		ChangeEventsTopic:    changeEventsTopic,
		AccessKey:            accessKeyFile,
		Certificate:          certificateFile,
		Ca:                   caFile,
	}
	generatedCode := c.GenerateM3ConnectorConsumer(zipFileName,
		"M3ConnectorConsumer",
		environment,
		"username",
		kafkaConfig)
	s.logContext.WithField("numberOfBytes", len(generatedCode)).Info("Code is generated")

	w.Write(generatedCode)
}
