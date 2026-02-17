package repository

// GitLabAccessConfig defines GitLab access permissions config.
type GitLabAccessConfig struct {
	// Group is the GitLab group in which the access token will be created.
	Group string `yaml:"group,omitempty"`
	// Scopes are the GitLab access scopes.
	Scopes []string `yaml:"scopes,omitempty"`
}
