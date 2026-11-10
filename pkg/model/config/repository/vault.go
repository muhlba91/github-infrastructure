package repository

// VaultAccessPermissionsConfig defines vault access permissions config.
type VaultAccessPermissionsConfig struct {
	// Enabled indicates whether vault access is enabled.
	Enabled *bool `yaml:"enabled"`
	// Address is the vault address.
	Address *string `yaml:"address,omitempty"`
	// AdditionalMounts defines additional vault mount access permissions config.
	AdditionalMounts []VaultAdditionalMountAccessPermissionsConfig `yaml:"additionalMounts,omitempty"`
	// AppRoles defines Vault AppRole roles owned by the repository, keyed by the role and policy name (global in Vault).
	AppRoles map[string]VaultAppRoleConfig `yaml:"appRoles,omitempty"`
	// TokenRoles defines Vault token roles (auth/token/roles) owned by the repository, keyed by the role name (global
	// in Vault). Who may create tokens through a role is decided by the paths of other policies
	// (auth/token/create/<role>), not here.
	TokenRoles map[string]VaultTokenRoleConfig `yaml:"tokenRoles,omitempty"`
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

// VaultAppRoleConfig defines a Vault AppRole role, and its policy, owned by the repository.
// The repository's CI may read the role ID and create secret IDs, but not change the role or its policy.
type VaultAppRoleConfig struct {
	// TokenTTL is the TTL of issued tokens in seconds (default: 1 hour).
	TokenTTL *int `yaml:"tokenTtl,omitempty"`
	// TokenMaxTTL is the maximum TTL of issued tokens in seconds (default: 24 hours).
	TokenMaxTTL *int `yaml:"tokenMaxTtl,omitempty"`
	// Paths defines the policy of the role: exact Vault paths (globs allowed) and their capabilities.
	Paths []VaultPathAccessPermissionsConfig `yaml:"paths"`
}

// VaultPathAccessPermissionsConfig defines the capabilities on an exact Vault path.
type VaultPathAccessPermissionsConfig struct {
	// Path is the Vault path, used as is (e.g. "github-repo/data/hosts/host/*").
	Path string `yaml:"path"`
	// Permissions defines the capabilities on the path.
	Permissions []string `yaml:"permissions"`
}

// VaultTokenRoleConfig defines a Vault token role: which policies a token created through it may carry.
type VaultTokenRoleConfig struct {
	// AllowedPolicies are exact policy names a token may carry.
	AllowedPolicies []string `yaml:"allowedPolicies,omitempty"`
	// AllowedPoliciesGlobs are policy name globs a token may carry (e.g. "github-*").
	AllowedPoliciesGlobs []string `yaml:"allowedPoliciesGlobs,omitempty"`
	// DisallowedPolicies are exact policy names a token may never carry (they win over the allowed ones).
	DisallowedPolicies []string `yaml:"disallowedPolicies,omitempty"`
	// DisallowedPoliciesGlobs are policy name globs a token may never carry (they win over the allowed ones).
	DisallowedPoliciesGlobs []string `yaml:"disallowedPoliciesGlobs,omitempty"`
	// Orphan creates tokens without a parent, so they outlive the token that created them (default: false).
	Orphan *bool `yaml:"orphan,omitempty"`
	// TokenTTL is the TTL of issued tokens in seconds (default: 1 hour).
	TokenTTL *int `yaml:"tokenTtl,omitempty"`
	// TokenMaxTTL is the maximum TTL of issued tokens in seconds (default: 4 hours).
	TokenMaxTTL *int `yaml:"tokenMaxTtl,omitempty"`
}
