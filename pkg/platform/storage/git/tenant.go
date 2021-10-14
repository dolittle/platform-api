package git

import (
	"path/filepath"
)

func (s *GitStorage) GetTenantDirectory(tenantID string) string {
	s.Pull()
	return filepath.Join(s.Directory, tenantID)
}
