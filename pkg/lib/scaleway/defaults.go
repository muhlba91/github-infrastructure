//nolint:gochecknoglobals // globals are allowed in this file
package scaleway

import "github.com/muhlba91/github-infrastructure/pkg/lib/config"

// maxApplicationNameLength defines the maximum length for Scaleway application names, as per Scaleway's API constraints.
const maxApplicationNameLength = 64

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

// commonLabels generates a list of common labels to be applied to all Scaleway resources, based on the common labels defined in the configuration.
func commonLabels() []string {
	labels := make([]string, 0, len(config.CommonLabels()))
	for k, v := range config.CommonLabels() {
		labels = append(labels, k+"="+v)
	}
	return labels
}
