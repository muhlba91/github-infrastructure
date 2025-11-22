//nolint:gochecknoglobals // globals are allowed in this file
package aws

// maxRepositoryLength defines the maximum length for the repository name used in AWS resource namings.
const maxRepositoryLength = 18

// postfixLength defines the length of the random postfix added to resource names for uniqueness.
const postfixLength = 8

// defaultPermissions defines the default set of AWS permissions assigned to service accounts.
var defaultPermissions = []string{
	"iam:*",
	"s3:*",
	"kms:*",
}
