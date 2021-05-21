package utils

func StringArrayContains(items []string, find string) bool {
	for _, item := range items {
		if item == find {
			return true
		}
	}
	return false
}
