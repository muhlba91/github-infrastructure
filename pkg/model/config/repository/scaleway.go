package repository

// ScalewayAccessConfig defines Scaleway access config.
type ScalewayAccessConfig struct {
	// OrganizationID is the ID of the Scaleway organization.
	OrganizationID *string `yaml:"organizationId,omitempty"`
	// Region is the cloud region.
	Region *string `yaml:"region,omitempty"`
	// Zone is the cloud zone.
	Zone *string `yaml:"zone,omitempty"`
	// Project is the Scaleway project ID.
	Project *string `yaml:"project"`
	// IAMPermissions defines the IAM permissions.
	IAMPermissions []string `yaml:"iamPermissions,omitempty"`
	// LinkedProjects defines the linked Scaleway projects.
	LinkedProjects map[string]ScalewayLinkedAccessConfig `yaml:"linkedProjects,omitempty"`
}

// ScalewayLinkedAccessConfig defines Scaleway linked access config.
type ScalewayLinkedAccessConfig struct {
	// AccessLevel is the access level for the linked project.
	AccessLevel string `yaml:"accessLevel"`
	// IAMPermissions defines the IAM permissions for the linked project.
	IAMPermissions []string `yaml:"iamPermissions,omitempty"`
}
