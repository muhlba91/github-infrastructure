package repository

// AwsAccessConfig defines AWS access config.
type AwsAccessConfig struct {
	// Region is the cloud region.
	Region *string `yaml:"region,omitempty"`
	// IAMPermissions defines the IAM permissions.
	IAMPermissions []string `yaml:"iamPermissions,omitempty"`
	// Account is the AWS account ID.
	Account *string `yaml:"account"`
}
