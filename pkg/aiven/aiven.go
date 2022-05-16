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
		logContext.WithField("error", err).Error("failed to create the aiven client with token")
		return nil, err
	}

	// the CA is unchanging so we can create it on initialization
	certificateAuthority, err := aivenClient.CA.Get(project)
	if err != nil {
		logContext.WithField("error", err).Error("failed to get the certificate authority")
		return nil, err
	}

	aivenService, err := aivenClient.Services.Get(project, service)
	if err != nil {
		logContext.WithField("error", err).Error("failed to get the services information")
		return nil, err
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

func (c *Client) CreateUser(username string) (string, string, error) {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":   "CreateUser",
		"username": username,
	})
	logContext.Debug("creating a user")
	userRequest := aiven.CreateServiceUserRequest{
		Username: username,
	}

	serviceUser, err := c.client.ServiceUsers.Create(c.project, c.service, userRequest)
	if err != nil {
		logContext.WithField("error", err).Error("failed to create the service user")
		return "", "", err
	}
	logContext.Debug("created the service user")
	return serviceUser.AccessCert, serviceUser.AccessKey, err
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
		logContext.WithField("error", err).Error("failed to create the ACL")
		return err
	}
	return err
}

func (c *Client) CreateTopic(topic string, retentionMs int64) error {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":       "CreateTopic",
		"topic":        topic,
		"retention_ms": retentionMs,
	})
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
		logContext.WithField("error", err).Error("failed to create the topic")
	}
	return err
}

func (c *Client) GetCertificateAuthority() string {
	return c.certificateAuthority
}

func (c *Client) GetBrokerUrl() string {
	return c.brokerUrl
}
