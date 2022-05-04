package aiven

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type client struct {
	apiToken string
	client   *http.Client
	project  string
	service  string
	aivenURL string
}

func NewClient(apiToken, project, service string) *client {
	return &client{
		apiToken: apiToken,
		project:  project,
		service:  service,
		client:   &http.Client{},
		aivenURL: "https://api.aiven.io/v1",
	}
}

type createUserRequest struct {
	AccessControl  accessControl `json:"access_control"`
	Authentication string        `json:"authentication,omitempty"`
	Username       string        `json:"username"`
}
type accessControl struct {
	M3Group            string   `json:"m3_group,omitempty"`
	PgAllowReplication bool     `json:"pg_allow_replication,omitempty"`
	RedisACLCategories []string `json:"redis_acl_categories,omitempty"`
	RedisACLChannels   []string `json:"redis_acl_channels,omitempty"`
	RedisACLCommands   []string `json:"redis_acl_commands,omitempty"`
	RedisACLKeys       []string `json:"redis_acl_keys,omitempty"`
}

func (c *client) CreateUser(username string) (createUserResponse, error) {
	var userResponse createUserResponse
	if username == "" {
		// bruh
		return userResponse, errors.New("empty usernames are not allowed")
	}

	url := fmt.Sprintf("%s/project/%s/service/%s/user", c.aivenURL, c.project, c.service)

	userRequest := createUserRequest{
		Username: username,
	}
	body, err := json.Marshal(userRequest)
	if err != nil {
		return userResponse, err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return userResponse, err
	}

	response, err := c.do(request)
	if err != nil {
		return userResponse, err
	}

	err = json.NewDecoder(response.Body).Decode(&userResponse)
	if err != nil {
		return userResponse, err
	}

	return userResponse, nil

}

type createUserResponse struct {
	Errors  []Errors `json:"errors"`
	Message string   `json:"message"`
	User    User     `json:"user"`
}
type Errors struct {
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}
type AccessControl struct {
	M3Group            string   `json:"m3_group"`
	PgAllowReplication bool     `json:"pg_allow_replication"`
	RedisACLCategories []string `json:"redis_acl_categories"`
	RedisACLChannels   []string `json:"redis_acl_channels"`
	RedisACLCommands   []string `json:"redis_acl_commands"`
	RedisACLKeys       []string `json:"redis_acl_keys"`
}
type User struct {
	AccessCert                    string        `json:"access_cert"`
	AccessCertNotValidAfterTime   string        `json:"access_cert_not_valid_after_time"`
	AccessControl                 AccessControl `json:"access_control"`
	AccessKey                     string        `json:"access_key"`
	Authentication                string        `json:"authentication"`
	ExpiringCertNotValidAfterTime string        `json:"expiring_cert_not_valid_after_time"`
	Password                      string        `json:"password"`
	Type                          string        `json:"type"`
	Username                      string        `json:"username"`
}

type KafkaACLPermission string

const (
	Admin     KafkaACLPermission = "admin"
	Read      KafkaACLPermission = "read"
	ReadWrite KafkaACLPermission = "readwrite"
	Write     KafkaACLPermission = "write"
)

type createACLRequest struct {
	Permission string `json:"permission"`
	Topic      string `json:"topic"`
	Username   string `json:"username"`
}

func (c *client) AddEnvironment(customer, application, environment string) error {

	return nil
}

func (c *client) CreateACL(topic string, username string, permission KafkaACLPermission) (createACLResponse, error) {
	var createACLResponse createACLResponse
	if topic == "" {
		return createACLResponse, errors.New("empty topics are not allowed")
	}
	if username == "" {
		// bruh
		return createACLResponse, errors.New("empty usernames are not allowed")
	}

	url := fmt.Sprintf("%s/project/%s/service/%s/acl", c.aivenURL, c.project, c.service)

	createACLRequest := createACLRequest{
		Permission: string(permission),
		Username:   username,
		Topic:      topic,
	}
	body, err := json.Marshal(createACLRequest)
	if err != nil {
		return createACLResponse, err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return createACLResponse, err
	}

	response, err := c.do(request)
	if err != nil {
		return createACLResponse, err
	}

	err = json.NewDecoder(response.Body).Decode(&createACLResponse)
	if err != nil {
		return createACLResponse, err
	}

	return createACLResponse, nil
}

type createACLResponse struct {
	ACL     []ACL    `json:"acl"`
	Errors  []Errors `json:"errors"`
	Message string   `json:"message"`
}
type ACL struct {
	ID         string `json:"id"`
	Permission string `json:"permission"`
	Topic      string `json:"topic"`
	Username   string `json:"username"`
}

type createTopicRequest struct {
	CleanupPolicy     string `json:"cleanup_policy"`
	Config            Config `json:"config"`
	MinInsyncReplicas int    `json:"min_insync_replicas"`
	Partitions        int    `json:"partitions"`
	Replication       int    `json:"replication"`
	RetentionBytes    int    `json:"retention_bytes"`
	RetentionHours    int64  `json:"retention_hours"`
	Tags              []Tags `json:"tags"`
	TopicName         string `json:"topic_name"`
}

type Config struct {
	CleanupPolicy                   string `json:"cleanup_policy"`
	CompressionType                 string `json:"compression_type"`
	DeleteRetentionMs               int    `json:"delete_retention_ms"`
	FileDeleteDelayMs               int    `json:"file_delete_delay_ms"`
	FlushMessages                   int    `json:"flush_messages"`
	FlushMs                         int    `json:"flush_ms"`
	IndexIntervalBytes              int    `json:"index_interval_bytes"`
	MaxCompactionLagMs              int    `json:"max_compaction_lag_ms"`
	MaxMessageBytes                 int    `json:"max_message_bytes"`
	MessageDownconversionEnable     bool   `json:"message_downconversion_enable"`
	MessageFormatVersion            string `json:"message_format_version"`
	MessageTimestampDifferenceMaxMs int    `json:"message_timestamp_difference_max_ms"`
	MessageTimestampType            string `json:"message_timestamp_type"`
	MinCleanableDirtyRatio          int    `json:"min_cleanable_dirty_ratio"`
	MinCompactionLagMs              int    `json:"min_compaction_lag_ms"`
	MinInsyncReplicas               int    `json:"min_insync_replicas"`
	Preallocate                     bool   `json:"preallocate"`
	RetentionBytes                  int    `json:"retention_bytes"`
	RetentionMs                     int    `json:"retention_ms"`
	SegmentBytes                    int    `json:"segment_bytes"`
	SegmentIndexBytes               int    `json:"segment_index_bytes"`
	SegmentJitterMs                 int    `json:"segment_jitter_ms"`
	SegmentMs                       int    `json:"segment_ms"`
	UncleanLeaderElectionEnable     bool   `json:"unclean_leader_election_enable"`
}
type Tags struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (c *client) CreateTopic(topic string, retentionMs int32) (createTopicResponse, error) {
	var createResponse createTopicResponse
	if topic == "" {
		return createResponse, errors.New("topic can't be empty")
	}

	url := fmt.Sprintf("%s/project/%s/service/%s/topic", c.aivenURL, c.project, c.service)

	createRequest := createTopicRequest{
		TopicName: topic,
		Config: Config{
			RetentionMs: int(retentionMs),
		},
	}
	body, err := json.Marshal(createRequest)
	if err != nil {
		return createResponse, err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return createResponse, err
	}

	response, err := c.do(request)
	if err != nil {
		return createResponse, err
	}

	err = json.NewDecoder(response.Body).Decode(&createResponse)
	if err != nil {
		return createResponse, err
	}

	return createResponse, nil
}

type createTopicResponse struct {
	Errors  []Errors `json:"errors"`
	Message string   `json:"message"`
}

func (c *client) do(request *http.Request) (response *http.Response, err error) {
	request.Header.Set("authorization", fmt.Sprintf("aivenv1 %s", c.apiToken))
	fmt.Println(request)
	return c.client.Do(request)
}
