package git

import (
	"path/filepath"
)

func (s *GitStorage) GetTenantDirectory(tenantID string) string {
	return filepath.Join(s.Directory, tenantID)
}
