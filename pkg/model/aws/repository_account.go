package aws

// RepositoryAccount defines an AWS account for a repository.
type RepositoryAccount struct {
	// Repository is the name of the repository.
	Repository *string
	// ID is the AWS account ID.
	ID *string
	// Region is the AWS region.
	Region *string
	// IAMPermissions are the IAM permissions for the repository.
	IAMPermissions []string
}
