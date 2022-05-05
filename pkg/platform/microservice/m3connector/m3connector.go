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
	logContext logrus.FieldLogger
}

type KafkaProvider interface {
	CreateTopic(topic string, retentionMs int64) error
	CreateUser(username string) error
	AddACL(topic string, username string, permission string) error
}

type KafkaACLPermission string

const (
	Admin     KafkaACLPermission = "admin"
	Read      KafkaACLPermission = "read"
	ReadWrite KafkaACLPermission = "readwrite"
	Write     KafkaACLPermission = "write"
)

const serviceName = "m3connector"

func NewM3Connector(kafka KafkaProvider, logContext logrus.FieldLogger) *M3Connector {
	return &M3Connector{
		kafka: kafka,
		logContext: logContext.WithFields(logrus.Fields{
			"context": "m3connector",
		}),
	}
}

// CreateEnvironment creates the required 4 topics, ACL's for them and an user needed for M3Connector to work in an environment
func (m *M3Connector) CreateEnvironment(customerID, applicationID, environment string) error {
	logContext := m.logContext.WithField("method", "CreateEnvironment")

	if customerID == "" {
		return errors.New("customer can't be empty")
	}
	if applicationID == "" {
		return errors.New("application can't be empty")
	}
	if environment == "" {
		return errors.New("environment can't be empty")
	}

	customerID = strings.ToLower(customerID)
	applicationID = strings.ToLower(applicationID)
	environment = strings.ToLower(environment)

	resourcePrefix := fmt.Sprintf("cust_%s.app_%s.env_%s.%s", customerID, applicationID, environment, serviceName)
	shortCustomerID := strings.ReplaceAll(customerID, "-", "")[:16]
	shortApplicationID := strings.ReplaceAll(applicationID, "-", "")[:16]
	username := fmt.Sprintf("%s.%s.%s.%s", shortCustomerID, shortApplicationID, environment, serviceName)

	logContext = logContext.WithFields(logrus.Fields{
		"customer_id":    customerID,
		"application_id": applicationID,
		"environment":    environment,
		"username":       username,
	})

	err := m.kafka.CreateUser(username)
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

	logContext.Debug("created all topics and ACL's")

	return nil
}

func (m *M3Connector) createTopicAndACL(topic string, retentionMs int64, username string) error {
	logContext := m.logContext.WithFields(logrus.Fields{
		"method":       "createTopicAndACL",
		"topic":        topic,
		"retention_ms": retentionMs,
		"username":     username,
	})
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
