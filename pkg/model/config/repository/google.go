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
}

// GoogleLinkedAccessConfig defines Google linked access config.
type GoogleLinkedAccessConfig struct {
	// AccessLevel is the access level for the linked project.
	AccessLevel string `yaml:"accessLevel"`
	// IAMPermissions defines the IAM permissions for the linked project.
	IAMPermissions []string `yaml:"iamPermissions,omitempty"`
}
