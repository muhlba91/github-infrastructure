package vault

// Config defines Vault-related configuration.
type Config struct {
	// Enabled indicates whether Vault integration is enabled.
	Enabled *bool `yaml:"enabled,omitempty"`
	// Address is the address of the Vault server.
	Address *string `yaml:"address,omitempty"`
}
