package m3connector

import (
	"errors"
	"fmt"
	"time"
)

type M3Connector struct {
	kafka KafkaProvider
}

type KafkaProvider interface {
	CreateTopic(topic string, retentionMs int64) error
	CreateUser(username string) error
	CreateACL(topic string, username string, permission string) error
}

func NewM3Connector(kafka KafkaProvider) *M3Connector {
	return &M3Connector{
		kafka: kafka,
	}
}

func (m *M3Connector) CreateEnvironment(customer, application, environment string) error {
	if customer == "" {
		return errors.New("customer can't be empty")
	}
	if application == "" {
		return errors.New("application can't be empty")
	}
	if environment == "" {
		return errors.New("environment can't be empty")
	}

	resourcePrefix := fmt.Sprintf("cust_%s_%s_%s.m3", customer, application, environment)
	username := resourcePrefix

	err := m.kafka.CreateUser(resourcePrefix)
	if err != nil {
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

	return nil
}

func (m *M3Connector) createTopicAndACL(topic string, retentionMs int64, username string) error {
	err := m.kafka.CreateTopic(topic, retentionMs)
	if err != nil {
		return err
	}

	err = m.kafka.CreateACL(topic, username, "readwrite")
	if err != nil {
		return err
	}
	return nil
}
