package m3connector

import (
	"errors"
	"fmt"
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

	username := fmt.Sprintf("cust_%s_%s_%s.m3", customer, application, environment)
	m.kafka.CreateUser(username)

	return nil
}
