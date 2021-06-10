package git

import (
	"fmt"
)

func (s *GitStorage) GetTenantDirectory(tenantID string) string {
	return fmt.Sprintf("%s/%s", s.Directory, tenantID)
}
