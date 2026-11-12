package repository

// AccessPermissionsConfig defines access permissions config.
type AccessPermissionsConfig struct {
	GitLab *GitLabAccessConfig `yaml:"gitlab,omitempty"`
	// Tailscale indicates whether to enable Tailscale access.
	Tailscale *bool `yaml:"tailscale,omitempty"`
	// Vault defines the vault access permissions config.
	Vault *VaultAccessPermissionsConfig `yaml:"vault,omitempty"`
	// GitHubSecrets defines whether secrets are written as GitHub Actions secrets of the repository.
	GitHubSecrets *GitHubSecretsConfig `yaml:"githubSecrets,omitempty"`
	// Google defines the Google cloud access config.
	Google *GoogleAccessConfig `yaml:"google,omitempty"`
	// Aws defines the AWS access config.
	Aws *AwsAccessConfig `yaml:"aws,omitempty"`
	// Scaleway defines the Scaleway access config.
	Scaleway *ScalewayAccessConfig `yaml:"scaleway,omitempty"`
}
