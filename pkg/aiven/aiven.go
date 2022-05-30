package aiven

import (
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/sirupsen/logrus"
)

type Client struct {
	client               *aiven.Client
	project              string
	service              string
	logContext           logrus.FieldLogger
	certificateAuthority string
	brokerUrl            string
}

// NewClient creates a new aiven token client with a default user agent string copied from their own implementation
// https://github.com/aiven/aiven-go-client/blob/31343720eb5c31fbe37fe2ac188daa435b99ee4c/client.go#L71
func NewClient(apiToken, project, service string, logger logrus.FieldLogger) (*Client, error) {
	userAgent := fmt.Sprintf("aiven-go-client/%s", aiven.Version())
	logContext := logger.WithFields(logrus.Fields{
		"context":    "aiven",
		"user_agent": userAgent,
	})
	aivenClient, err := aiven.NewTokenClient(apiToken, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create an aiven client with token: %w", err)
	}

	// the CA is unchanging so we can create it on initialization
	certificateAuthority, err := aivenClient.CA.Get(project)
	if err != nil {
		return nil, fmt.Errorf("failed to get the certificate authority: %w", err)
	}

	aivenService, err := aivenClient.Services.Get(project, service)
	if err != nil {
		return nil, fmt.Errorf("failed to get the services information: %w", err)
	}

	return &Client{
		project:              project,
		service:              service,
		client:               aivenClient,
		certificateAuthority: certificateAuthority,
		logContext:           logContext,
		brokerUrl:            aivenService.URI,
	}, nil
}

func (c *Client) GetOrCreateUser(username string) (string, string, error) {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":   "GetOrCreateUser",
		"username": username,
	})
	var serviceUser *aiven.ServiceUser

	logContext.Debug("getting the service user user")
	serviceUser, getErr := c.client.ServiceUsers.Get(c.project, c.service, username)
	if getErr != nil {
		if !aiven.IsNotFound(getErr) {
			return "", "", fmt.Errorf("failed to check if the user already existed: %w", getErr)
		}

		logContext.Debug("no existing user already found, creating a new service user")
		userRequest := aiven.CreateServiceUserRequest{
			Username: username,
		}

		// have to initialize createErr so that we can use '=' to mutate the serviceUser variable from outer scope
		var createErr error
		serviceUser, createErr = c.client.ServiceUsers.Create(c.project, c.service, userRequest)
		if createErr != nil {
			return "", "", fmt.Errorf("failed to create a user: %w", createErr)
		}
		logContext.Debug("created the service user")
	}

	return serviceUser.AccessCert, serviceUser.AccessKey, nil
}

func (c *Client) AddACL(topic string, username string, permission string) error {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":     "CreateACL",
		"username":   username,
		"topic":      topic,
		"permission": permission,
	})

	logContext.Debug("adding an acl")
	userRequest := aiven.CreateKafkaACLRequest{
		Permission: permission,
		Topic:      topic,
		Username:   username,
	}

	_, err := c.client.KafkaACLs.Create(c.project, c.service, userRequest)
	if err != nil {
		if !aiven.IsAlreadyExists(err) {
			return fmt.Errorf("failed to add an ACl: %w", err)
		}
		logContext.Debug("the ACL already exists")
	}
	return nil
}

func (c *Client) CreateTopic(topic string, retentionMs int64) error {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":       "CreateTopic",
		"topic":        topic,
		"retention_ms": retentionMs,
	})
	logContext.Debug("creating a topic")
	replication := 3
	topicRequest := aiven.CreateKafkaTopicRequest{
		TopicName:   topic,
		Replication: &replication,
		Config: aiven.KafkaTopicConfig{
			RetentionMs: &retentionMs,
		},
	}

	err := c.client.KafkaTopics.Create(c.project, c.service, topicRequest)
	if err != nil {
		if !aiven.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create a topic: %w", err)
		}
		logContext.Debug("the topic already exists")
	}
	return nil
}

func (c *Client) GetCertificateAuthority() string {
	return c.certificateAuthority
}

func (c *Client) GetBrokerUrl() string {
	return c.brokerUrl
}
