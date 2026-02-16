package repository

// AccessPermissionsConfig defines access permissions config.
type AccessPermissionsConfig struct {
	// Tailscale indicates whether to enable Tailscale access.
	Tailscale *bool `yaml:"tailscale,omitempty"`
	// Vault defines the vault access permissions config.
	Vault *VaultAccessPermissionsConfig `yaml:"vault,omitempty"`
	// Google defines the Google cloud access config.
	Google *GoogleAccessConfig `yaml:"google,omitempty"`
	// Aws defines the AWS access config.
	Aws *AwsAccessConfig `yaml:"aws,omitempty"`
	// Scaleway defines the Scaleway access config.
	Scaleway *ScalewayAccessConfig `yaml:"scaleway,omitempty"`
}
