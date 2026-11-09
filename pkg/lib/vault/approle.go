package vault

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/approle"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/policy"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

const (
	// appRoleBackend is the mount path of the AppRole auth method.
	appRoleBackend = "approle"
	// policyPathKey is the template key of a path in a Vault policy.
	policyPathKey = "path"
	// policyPermissionsKey is the template key of the capabilities of a path in a Vault policy.
	policyPermissionsKey = "permissions"
)

// appRoles returns the AppRole roles of the given repository and their names, sorted.
// The sorted names keep rendered policies stable across runs (map iteration order is random).
// repository: The repository configuration.
func appRoles(repository *repository.Config) (map[string]repository.VaultAppRoleConfig, []string) {
	if repository.AccessPermissions == nil || repository.AccessPermissions.Vault == nil {
		return nil, nil
	}
	roles := repository.AccessPermissions.Vault.AppRoles
	return roles, slices.Sorted(maps.Keys(roles))
}

// createAppRoles creates the AppRole roles of the given repository, each with a policy of the same name.
// The repository's CI only gets the rights to read the role IDs and manage secret IDs (see createPolicy),
// so it cannot change what a role may access.
// ctx: The Pulumi context.
// repository: The repository configuration.
func createAppRoles(ctx *pulumi.Context, repository *repository.Config) error {
	roles, names := appRoles(repository)
	for _, name := range names {
		role := roles[name]
		paths := []map[string]string{}
		for _, path := range role.Paths {
			paths = append(paths, map[string]string{
				policyPathKey:        path.Path,
				policyPermissionsKey: quotePermissions(path.Permissions),
			})
		}

		rendered, rErr := template.Render("assets/vault/approle-policy.hcl.tpl", map[string]any{
			"paths": paths,
		})
		if rErr != nil {
			log.Err(rErr).Msgf("[vault][approle] error rendering Vault policy for AppRole: %s", name)
			return rErr
		}

		rolePolicy, pErr := policy.Create(ctx, &policy.CreateOptions{
			Name:          name,
			Policy:        pulumi.String(rendered),
			PulumiOptions: []pulumi.ResourceOption{pulumi.Provider(config.VaultProvider)},
		})
		if pErr != nil {
			log.Err(pErr).Msgf("[vault][approle] error creating Vault policy for AppRole: %s", name)
			return pErr
		}

		backend := appRoleBackend
		_, aErr := approle.Create(ctx, name, &approle.CreateOptions{
			Backend:       &backend,
			TokenPolicies: pulumi.ToStringArray([]string{name}),
			TokenTTL:      role.TokenTTL,
			TokenMaxTTL:   role.TokenMaxTTL,
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(config.VaultProvider),
				pulumi.DependsOn([]pulumi.Resource{rolePolicy}),
			},
		})
		if aErr != nil {
			log.Err(aErr).Msgf("[vault][approle] error creating Vault AppRole: %s", name)
			return aErr
		}
	}

	return nil
}

// quotePermissions renders a list of capabilities for a Vault policy (e.g. `"read", "list"`).
// permissions: The capabilities.
func quotePermissions(permissions []string) string {
	quoted := []string{}
	for _, permission := range permissions {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, permission))
	}
	return strings.Join(quoted, ", ")
}
