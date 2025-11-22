package repository

// VaultAccessPermissionsConfig defines vault access permissions config.
type VaultAccessPermissionsConfig struct {
	// Enabled indicates whether vault access is enabled.
	Enabled *bool `yaml:"enabled"`
	// Address is the vault address.
	Address *string `yaml:"address,omitempty"`
	// AdditionalMounts defines additional vault mount access permissions config.
	AdditionalMounts []VaultAdditionalMountAccessPermissionsConfig `yaml:"additionalMounts,omitempty"`
}

// VaultAdditionalMountAccessPermissionsConfig defines vault additional mount access permissions config.
type VaultAdditionalMountAccessPermissionsConfig struct {
	// Path is the vault mount path.
	Path string `yaml:"path"`
	// Create indicates whether to create the mount.
	Create *bool `yaml:"create,omitempty"`
	// Permissions defines the permissions for the mount.
	Permissions []string `yaml:"permissions"`
}
