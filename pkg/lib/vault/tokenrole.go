package vault

import (
	"fmt"
	"maps"
	"slices"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault/tokenauth"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

const (
	// defaultTokenRoleTTL is the default TTL of tokens created through a token role (1 hour).
	defaultTokenRoleTTL = 60 * 60
	// defaultTokenRoleMaxTTL is the default maximum TTL of tokens created through a token role (4 hours).
	defaultTokenRoleMaxTTL = 4 * 60 * 60
)

// createTokenRoles creates the token roles of the given repository.
// A token role only limits the policies of the tokens created through it; who may use it is decided by other
// policies granting auth/token/create/<role> (e.g. an AppRole's paths).
// ctx: The Pulumi context.
// repository: The repository configuration.
func createTokenRoles(ctx *pulumi.Context, repository *repository.Config) error {
	if repository.AccessPermissions == nil || repository.AccessPermissions.Vault == nil {
		return nil
	}
	roles := repository.AccessPermissions.Vault.TokenRoles

	for _, name := range slices.Sorted(maps.Keys(roles)) {
		role := roles[name]
		if len(role.AllowedPolicies) == 0 && len(role.AllowedPoliciesGlobs) == 0 {
			err := fmt.Errorf("token role %s allows no policies", name)
			log.Err(err).Msgf("[vault][tokenrole] invalid Vault token role: %s", name)
			return err
		}

		_, err := tokenauth.NewAuthBackendRole(
			ctx,
			fmt.Sprintf("vault-token-role-%s", name),
			&tokenauth.AuthBackendRoleArgs{
				RoleName:                pulumi.String(name),
				AllowedPolicies:         pulumi.ToStringArray(role.AllowedPolicies),
				AllowedPoliciesGlobs:    pulumi.ToStringArray(role.AllowedPoliciesGlobs),
				DisallowedPolicies:      pulumi.ToStringArray(role.DisallowedPolicies),
				DisallowedPoliciesGlobs: pulumi.ToStringArray(role.DisallowedPoliciesGlobs),
				Orphan:                  pulumi.Bool(defaults.GetOrDefault(role.Orphan, false)),
				Renewable:               pulumi.Bool(true),
				TokenTtl:                pulumi.Int(defaults.GetOrDefault(role.TokenTTL, defaultTokenRoleTTL)),
				TokenMaxTtl:             pulumi.Int(defaults.GetOrDefault(role.TokenMaxTTL, defaultTokenRoleMaxTTL)),
			},
			pulumi.Provider(config.VaultProvider),
		)
		if err != nil {
			log.Err(err).Msgf("[vault][tokenrole] error creating Vault token role: %s", name)
			return err
		}
	}

	return nil
}
