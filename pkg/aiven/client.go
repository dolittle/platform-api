package aiven

import (
	"errors"
	"fmt"
	"log"

	"github.com/aiven/aiven-go-client"
)

type client struct {
	client  *aiven.Client
	project string
	service string
}

func NewClient(apiToken, project, service string) *client {
	// the  user agent string is by default like that, I just wanted to be explicit
	// https://github.com/aiven/aiven-go-client/blob/31343720eb5c31fbe37fe2ac188daa435b99ee4c/client.go#L71
	aivenClient, err := aiven.NewTokenClient(apiToken, fmt.Sprintf("aiven-go-client/%s", aiven.Version()))
	if err != nil {
		// something
		log.Fatalf("error creating token client: %s", err)
	}
	return &client{
		project: project,
		service: service,
		client:  aivenClient,
	}
}

func (c *client) CreateUser(username string) (*aiven.ServiceUser, error) {
	if username == "" {
		return nil, errors.New("empty usernames are not allowed")
	}

	userRequest := aiven.CreateServiceUserRequest{
		Username: username,
	}

	response, err := c.client.ServiceUsers.Create(c.project, c.service, userRequest)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *client) AddEnvironment(customer, application, environment string) error {

	return nil
}

type KafkaACLPermission string

const (
	Admin     KafkaACLPermission = "admin"
	Read      KafkaACLPermission = "read"
	ReadWrite KafkaACLPermission = "readwrite"
	Write     KafkaACLPermission = "write"
)

func (c *client) CreateACL(topic string, username string, permission KafkaACLPermission) (*aiven.KafkaACL, error) {
	if topic == "" {
		return nil, errors.New("empty topics are not allowed")
	}
	if username == "" {
		return nil, errors.New("empty usernames are not allowed")
	}

	userRequest := aiven.CreateKafkaACLRequest{
		Permission: string(permission),
		Topic:      topic,
		Username:   username,
	}

	response, err := c.client.KafkaACLs.Create(c.project, c.service, userRequest)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// just returns an error because KafkaTopics.Create also only returns an error weirdly
func (c *client) CreateTopic(topic string, retentionMs int64) error {
	if topic == "" {
		return errors.New("empty topics are not allowed")
	}

	topicRequest := aiven.CreateKafkaTopicRequest{
		TopicName: topic,
		Config: aiven.KafkaTopicConfig{
			RetentionMs: &retentionMs,
		},
	}

	err := c.client.KafkaTopics.Create(c.project, c.service, topicRequest)
	if err != nil {
		return err
	}
	return nil
}
