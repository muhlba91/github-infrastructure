package aws

// Config defines AWS-related configuration.
type Config struct {
	// DefaultRegion contains configuration for the default AWS region.
	DefaultRegion *string `yaml:"defaultRegion,omitempty"`
	// Account contains configuration for specific AWS accounts.
	Account map[string]*Account `yaml:"account,omitempty"`
}

// Account defines configuration for a specific AWS account.
type Account struct {
	// ExternalID is the external ID used for cross-account access.
	ExternalID *string `yaml:"externalId,omitempty"`
	// RoleARN is the ARN of the role to assume in the target account.
	RoleARN *string `yaml:"roleArn,omitempty"`
}
