//nolint:gochecknoglobals // globals are allowed in this file
package scaleway

// defaultOrganizationPermissions defines the default set of Scaleway permissions assigned to service accounts for the entire organization.
var defaultOrganizationPermissions = []string{
	"ProjectReadOnly",
	"IAMApplicationManager",
	"IAMGroupManager",
	"IAMPolicyManager",
	"OrganizationReadOnly",
}

// defaultProjectPermissions defines the default set of Scaleway permissions assigned to service accounts for projects.
var defaultProjectPermissions = []string{
	"ObjectStorageFullAccess",
	"SecretManagerFullAccess",
	"KeyManagerFullAccess",
}
