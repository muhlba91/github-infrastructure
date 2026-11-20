package aws

// Config defines AWS-related configuration.
type Config struct {
	// DefaultRegion contains configuration for the default AWS region.
	DefaultRegion *string `yaml:"defaultRegion,omitempty"`
	// Account contains configuration for specific AWS accounts.
	Account map[string]*Account `yaml:"account,omitempty"`
	// SSO contains the IAM Identity Center permission set the CI roles trust for local runs (nil: not trusted).
	SSO *SSO `yaml:"sso,omitempty"`
}

// SSO defines the IAM Identity Center permission set that may assume the CI roles.
type SSO struct {
	// OrganizationID is the ID of the AWS Organization the permission set's accounts belong to (required).
	OrganizationID *string `yaml:"organizationId,omitempty"`
	// PermissionSetName is the name of the permission set, defaults to ci-assume.
	PermissionSetName *string `yaml:"permissionSetName,omitempty"`
}

// Account defines configuration for a specific AWS account.
type Account struct {
	// ExternalID is the external ID used for cross-account access.
	ExternalID *string `yaml:"externalId,omitempty"`
	// RoleARN is the ARN of the role to assume in the target account.
	RoleARN *string `yaml:"roleArn,omitempty"`
}
