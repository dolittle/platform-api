package m3connector

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type M3Connector struct {
	kafka      KafkaProvider
	k8sRepo    K8sRepo
	logContext logrus.FieldLogger
}

type KafkaProvider interface {
	CreateTopic(topic string, retentionMs int64) error
	// Get's or create's a Kafka "user" and returns the access certificate and access key if successful
	GetOrCreateUser(username string) (certificate string, key string, err error)
	AddACL(topic string, username string, permission string) error
	GetCertificateAuthority() string
	GetBrokerUrl() string
}

type K8sRepo interface {
	UpsertKafkaFiles(applicationID, environment string, kafkaFiles KafkaFiles) error
}

type KafkaACLPermission string

const (
	Admin     KafkaACLPermission = "admin"
	Read      KafkaACLPermission = "read"
	ReadWrite KafkaACLPermission = "readwrite"
	Write     KafkaACLPermission = "write"
)

const serviceName = "m3connector"

type KafkaFiles struct {
	AccessKey            string      `json:"accessKey.pem"`
	CertificateAuthority string      `json:"ca.pem"`
	Certificate          string      `json:"certificate.pem"`
	Config               KafkaConfig `json:"config.json"`
}

type KafkaConfig struct {
	BrokerUrl string   `json:"brokerUrl"`
	Topics    []string `json:"topics"`
}

func NewM3Connector(kafka KafkaProvider, k8sRepo K8sRepo, logContext logrus.FieldLogger) *M3Connector {
	return &M3Connector{
		kafka:   kafka,
		k8sRepo: k8sRepo,
		logContext: logContext.WithFields(logrus.Fields{
			"context": "m3connector",
		}),
	}
}

// CreateEnvironment creates the required 4 topics, ACL's for them and an user needed for M3Connector to work in an environment
func (m *M3Connector) CreateEnvironment(customerID, applicationID, environment string) error {
	customerID = strings.ToLower(customerID)
	applicationID = strings.ToLower(applicationID)
	environment = strings.ToLower(environment)

	logContext := m.logContext.WithFields(logrus.Fields{
		"customer_id":    customerID,
		"application_id": applicationID,
		"environment":    environment,
		"method":         "CreateEnvironment",
	})
	logContext.Info("creating the environment for m3connector")

	if customerID == "" {
		return errors.New("customer can't be empty")
	}
	if applicationID == "" {
		return errors.New("application can't be empty")
	}
	if environment == "" {
		return errors.New("environment can't be empty")
	}

	resourcePrefix := fmt.Sprintf("cust_%s.app_%s.env_%s.%s", customerID, applicationID, environment, serviceName)
	// Aiven only allows 64 characters for the username so we do some truncating
	shortCustomerID := strings.ReplaceAll(customerID, "-", "")[:16]
	shortApplicationID := strings.ReplaceAll(applicationID, "-", "")[:16]
	username := fmt.Sprintf("%s.%s.%s.%s", shortCustomerID, shortApplicationID, environment, serviceName)
	logContext = logContext.WithField("username", username)

	certificate, accessKey, err := m.kafka.GetOrCreateUser(username)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to create a user")
		return err
	}

	changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
	inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
	commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
	receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)

	err = m.createTopicAndACL(changeTopic, int64(-1), username)
	if err != nil {
		return err
	}
	err = m.createTopicAndACL(inputTopic, int64(-1), username)
	if err != nil {
		return err
	}
	err = m.createTopicAndACL(commandTopic, int64(-1), username)
	if err != nil {
		return err
	}
	hours, _ := time.ParseDuration("168h")
	err = m.createTopicAndACL(receiptsTopic, hours.Milliseconds(), username)
	if err != nil {
		return err
	}

	kafkaFiles := KafkaFiles{
		AccessKey:            accessKey,
		Certificate:          certificate,
		CertificateAuthority: m.kafka.GetCertificateAuthority(),
		Config: KafkaConfig{
			BrokerUrl: m.kafka.GetBrokerUrl(),
			Topics: []string{
				changeTopic,
				inputTopic,
				commandTopic,
				receiptsTopic,
			},
		},
	}
	err = m.k8sRepo.UpsertKafkaFiles(applicationID, environment, kafkaFiles)
	if err != nil {
		return err
	}

	logContext.Info("finished creating the environment")

	return nil
}

func (m *M3Connector) createTopicAndACL(topic string, retentionMs int64, username string) error {
	logContext := m.logContext.WithFields(logrus.Fields{
		"method":       "createTopicAndACL",
		"topic":        topic,
		"retention_ms": retentionMs,
		"username":     username,
	})
	logContext.Debug("creating the topic and ACL")
	err := m.kafka.CreateTopic(topic, retentionMs)
	if err != nil {
		logContext.WithField("error", err).Error("failed to create the topic")
		return err
	}

	err = m.kafka.AddACL(topic, username, string(ReadWrite))
	if err != nil {
		logContext.WithField("error", err).Error("failed to create the ACL")
		return err
	}
	return nil
}
