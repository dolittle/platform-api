package aiven

import (
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/sirupsen/logrus"
)

type Client struct {
	client     *aiven.Client
	project    string
	service    string
	logContext logrus.FieldLogger
}

func NewClient(apiToken, project, service string, logger logrus.FieldLogger) (*Client, error) {
	// the  user agent string is by default like that, I just wanted to be explicit
	// https://github.com/aiven/aiven-go-client/blob/31343720eb5c31fbe37fe2ac188daa435b99ee4c/client.go#L71
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
	return &Client{
		project:    project,
		service:    service,
		client:     aivenClient,
		logContext: logContext,
	}, nil
}

func (c *Client) CreateUser(username string) error {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":   "CreateUser",
		"username": username,
	})
	userRequest := aiven.CreateServiceUserRequest{
		Username: username,
	}

	_, err := c.client.ServiceUsers.Create(c.project, c.service, userRequest)
	if err != nil {
		logContext.WithField("error", err).Error("failed to create the service user")
		return err
	}
	logContext.Debug("created the service user")
	return err
}

// CreateACL adds an ACL entry with the permission for the given topic and  username
func (c *Client) CreateACL(topic string, username string, permission string) error {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":     "CreateACL",
		"username":   username,
		"topic":      topic,
		"permission": permission,
	})
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

// just returns an error because KafkaTopics.Create also only returns an error weirdly
func (c *Client) CreateTopic(topic string, retentionMs int64) error {
	logContext := c.logContext.WithFields(logrus.Fields{
		"method":       "CreateTopic",
		"topic":        topic,
		"retention_ms": retentionMs,
	})
	topicRequest := aiven.CreateKafkaTopicRequest{
		TopicName: topic,
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
