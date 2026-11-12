package repository

// GoogleAccessConfig defines Google access config.
type GoogleAccessConfig struct {
	// Region is the cloud region.
	Region *string `yaml:"region,omitempty"`
	// IAMPermissions defines the IAM permissions.
	IAMPermissions []string `yaml:"iamPermissions,omitempty"`
	// Project is the Google Cloud project ID.
	Project *string `yaml:"project"`
	// LinkedProjects defines the linked Google projects.
	LinkedProjects map[string]GoogleLinkedAccessConfig `yaml:"linkedProjects,omitempty"`
	// EnabledServices defines the enabled Google services.
	EnabledServices []string `yaml:"enabledServices,omitempty"`
	// HMACKey indicates whether to enable HMAC key.
	HMACKey *bool `yaml:"hmacKey,omitempty"`
	// SecretManager grants the CI service account access to the repository's secrets in the default project's Secret
	// Manager: the secrets named <repository>--<NAME>.
	SecretManager *bool `yaml:"secretManager,omitempty"`
	// ServiceAccounts defines additional service accounts in the default project, keyed by the account ID.
	ServiceAccounts map[string]GoogleServiceAccountConfig `yaml:"serviceAccounts,omitempty"`
}

// GoogleServiceAccountConfig defines a service account with predefined roles and a key stored in Vault.
type GoogleServiceAccountConfig struct {
	// Roles maps a project ID (one of google.projects) to the predefined roles the account gets in it.
	Roles map[string][]string `yaml:"roles,omitempty"`
	// KeyRotationDays rotates the key every this many days (default: never).
	KeyRotationDays *int `yaml:"keyRotationDays,omitempty"`
	// Vault defines where the account's key is stored.
	Vault *GoogleServiceAccountVaultConfig `yaml:"vault,omitempty"`
}

// GoogleServiceAccountVaultConfig defines where a service account's key is stored in Vault.
type GoogleServiceAccountVaultConfig struct {
	// Mount is the KV mount (default: the repository's mount github-<repository>); the key is stored as
	// gcp-<account ID>, field "credentials" (the key file's JSON).
	Mount *string `yaml:"mount,omitempty"`
}

// GoogleLinkedAccessConfig defines Google linked access config.
type GoogleLinkedAccessConfig struct {
	// AccessLevel is the access level for the linked project.
	AccessLevel string `yaml:"accessLevel"`
	// IAMPermissions defines the IAM permissions for the linked project.
	IAMPermissions []string `yaml:"iamPermissions,omitempty"`
}
