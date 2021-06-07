package rawdatalog

type RawMomentMetadata struct {
	TenantID      string            `json:"tenantId"`
	ApplicationID string            `json:"applicationId"`
	Environment   string            `json:"environment"`
	Labels        map[string]string `json:"labels"`
}

type RawMoment struct {
	Kind     string            `json:"kind"`
	When     int64             `json:"when"`
	Metadata RawMomentMetadata `json:"metadata"`
	Data     interface{}       `json:"data"`
}

type Repo interface {
	Write(topic string, moment RawMoment) error
}
