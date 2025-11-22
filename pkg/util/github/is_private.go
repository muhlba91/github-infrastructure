package github

// IsPrivate checks if a repository is private.
// visibility: the visibility level to check.
func IsPrivate(visibility *string) bool {
	return visibility != nil && *visibility == "private"
}
