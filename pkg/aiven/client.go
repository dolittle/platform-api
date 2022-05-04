package aiven

import (
	"errors"
	"fmt"
	"log"

	"github.com/aiven/aiven-go-client"
)

type Client struct {
	client  *aiven.Client
	project string
	service string
}

func NewClient(apiToken, project, service string) *Client {
	// the  user agent string is by default like that, I just wanted to be explicit
	// https://github.com/aiven/aiven-go-client/blob/31343720eb5c31fbe37fe2ac188daa435b99ee4c/client.go#L71
	aivenClient, err := aiven.NewTokenClient(apiToken, fmt.Sprintf("aiven-go-client/%s", aiven.Version()))
	if err != nil {
		// something
		log.Fatalf("error creating token client: %s", err)
	}
	return &Client{
		project: project,
		service: service,
		client:  aivenClient,
	}
}

func (c *Client) CreateUser(username string) error {
	if username == "" {
		return errors.New("empty usernames are not allowed")
	}

	userRequest := aiven.CreateServiceUserRequest{
		Username: username,
	}

	_, err := c.client.ServiceUsers.Create(c.project, c.service, userRequest)
	return err
}

type KafkaACLPermission string

const (
	Admin     KafkaACLPermission = "admin"
	Read      KafkaACLPermission = "read"
	ReadWrite KafkaACLPermission = "readwrite"
	Write     KafkaACLPermission = "write"
)

func (c *Client) CreateACL(topic string, username string, permission string) error {
	if topic == "" {
		return errors.New("empty topics are not allowed")
	}
	if username == "" {
		return errors.New("empty usernames are not allowed")
	}
	aclPermission := KafkaACLPermission(permission)

	userRequest := aiven.CreateKafkaACLRequest{
		Permission: string(aclPermission),
		Topic:      topic,
		Username:   username,
	}

	_, err := c.client.KafkaACLs.Create(c.project, c.service, userRequest)
	return err
}

// just returns an error because KafkaTopics.Create also only returns an error weirdly
func (c *Client) CreateTopic(topic string, retentionMs int64) error {
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
	return err
}
