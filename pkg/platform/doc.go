package platform

type GitRepo interface {
	Write(tenantID string, applicationID string, data []byte) error
	Read(tenantID string, applicationID string) ([]byte, error)
	GetAll(tenantID string) ([]Application, error)
}
