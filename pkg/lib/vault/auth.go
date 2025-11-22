package vault

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	vaultConf "github.com/muhlba91/github-infrastructure/pkg/model/config/vault"
	ghSecret "github.com/muhlba91/pulumi-shared-library/pkg/lib/github/actions/secret"
	vaultSecret "github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault/jwt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// defaultTokenTTL is the default time-to-live for Vault tokens issued to GitHub repositories.
const defaultTokenTTL = 1 * 60 * 60

// createAuth creates a JWT authentication backend role in Vault for the given GitHub repository.
// ctx: The Pulumi context.
// repository: The repository configuration.
// mount: The Vault mount where the auth backend is enabled.
// githubRepository: The GitHub repository resource.
// repositoriesConfig: The overall repositories configuration.
// vaultConfig: The Vault configuration.
func createAuth(
	ctx *pulumi.Context,
	repository *repository.Config,
	mount *vault.Mount,
	githubRepository *github.Repository,
	repositoriesConfig *repositories.Config,
	vaultConfig *vaultConf.Config,
) (*jwt.AuthBackendRole, error) {
	perr := createPolicy(ctx, repository)
	if perr != nil {
		return nil, perr
	}

	vaultAddr := repository.AccessPermissions.Vault.Address
	if vaultAddr == nil || *vaultAddr == "" {
		vaultAddr = vaultConfig.Address
	}

	jwtRole, abrErr := jwt.NewAuthBackendRole(
		ctx,
		fmt.Sprintf("vault-jwt-github-role-%s", repository.Name),
		&jwt.AuthBackendRoleArgs{
			Backend:  pulumi.String("github"),
			RoleType: pulumi.String("jwt"),
			RoleName: pulumi.String(fmt.Sprintf("github-%s", repository.Name)),
			TokenPolicies: pulumi.StringArray{
				pulumi.String(fmt.Sprintf("github-%s", repository.Name)),
			},
			TokenTtl: pulumi.Int(defaultTokenTTL),
			BoundAudiences: pulumi.StringArray{
				pulumi.String(fmt.Sprintf("https://github.com/%s", *repositoriesConfig.Owner)),
			},
			UserClaim: pulumi.String("repository"),
			BoundClaims: pulumi.StringMap{
				"repository": pulumi.String(fmt.Sprintf("%s/%s", *repositoriesConfig.Owner, repository.Name)),
			},
		},
		pulumi.Provider(config.VaultProvider),
	)
	if abrErr != nil {
		return nil, abrErr
	}

	createSecrets(ctx, mount, jwtRole, githubRepository, vaultAddr)

	return jwtRole, nil
}

// createPolicy creates a Vault policy for the given GitHub repository.
// ctx: The Pulumi context.
// repository: The repository configuration.
func createPolicy(ctx *pulumi.Context, repository *repository.Config) error {
	additionalPaths := []map[string]string{}
	if repository.AccessPermissions != nil && repository.AccessPermissions.Vault != nil &&
		repository.AccessPermissions.Vault.AdditionalMounts != nil {
		for _, mount := range repository.AccessPermissions.Vault.AdditionalMounts {
			permissions := []string{}
			for _, permission := range mount.Permissions {
				permissions = append(permissions, fmt.Sprintf(`"%s"`, permission))
			}
			additionalPaths = append(additionalPaths, map[string]string{
				"path":        mount.Path,
				"permissions": strings.Join(permissions, ", "),
			})
		}
	}

	policy, prError := template.Render("assets/vault/policy.hcl.tpl", map[string]any{
		"repository":      repository.Name,
		"additionalPaths": additionalPaths,
	})
	if prError != nil {
		return prError
	}

	_, polErr := vault.NewPolicy(ctx, fmt.Sprintf("vault-policy-github-%s", repository.Name), &vault.PolicyArgs{
		Name:   pulumi.String(fmt.Sprintf("github-%s", repository.Name)),
		Policy: pulumi.String(policy),
	}, pulumi.Provider(config.VaultProvider))
	if polErr != nil {
		return polErr
	}

	return nil
}

// createSecrets creates the necessary secrets in Vault and GitHub Actions for the given repository.
// ctx: The Pulumi context.
// mount: The Vault mount where the auth backend is enabled.
// jwtRole: The JWT authentication backend role in Vault.
// githubRepository: The GitHub repository resource.
// vaultAddr: The address of the Vault server.
func createSecrets(
	ctx *pulumi.Context,
	mount *vault.Mount,
	jwtRole *jwt.AuthBackendRole,
	githubRepository *github.Repository,
	vaultAddr *string,
) {
	pulumi.All(mount.Path, jwtRole.RoleName).ApplyT(func(all []any) error {
		path, _ := all[0].(string)
		roleName, _ := all[1].(string)

		value, _ := json.Marshal(map[string]string{
			"address": *vaultAddr,
			"role":    roleName,
			"path":    "github",
		})

		_, _ = vaultSecret.Write(ctx, &vaultSecret.WriteArgs{
			Path:  path,
			Key:   "vault",
			Value: pulumi.String(value),
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(config.VaultProvider),
				pulumi.DependsOn([]pulumi.Resource{
					mount,
				}),
			},
		})

		return nil
	})

	ghSecret.Write(ctx, &ghSecret.WriteArgs{
		Key:        "VAULT_ADDR",
		Value:      pulumi.String(*vaultAddr),
		Repository: githubRepository,
	})
	ghSecret.Write(ctx, &ghSecret.WriteArgs{
		Key:        "VAULT_ROLE",
		Value:      jwtRole.RoleName,
		Repository: githubRepository,
	})
	ghSecret.Write(ctx, &ghSecret.WriteArgs{
		Key:        "VAULT_PATH",
		Value:      pulumi.String("github"),
		Repository: githubRepository,
	})
}
