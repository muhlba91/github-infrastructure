package repository

// GitHubSecretsConfig defines whether the repository's secrets are written as GitHub Actions secrets.
type GitHubSecretsConfig struct {
	// Enabled indicates whether the secrets are written as GitHub Actions secrets of the repository (default: false).
	Enabled *bool `yaml:"enabled"`
}
