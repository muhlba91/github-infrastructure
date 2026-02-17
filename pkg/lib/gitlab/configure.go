package gitlab

import (
	"encoding/json"
	"maps"
	"slices"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/gitlab/groupaccesstoken"
	vaultSecret "github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// Configure sets up GitLab configurations for the specified repositories.
// ctx: The Pulumi context for resource management.
// repositories: A slice of repository configurations.
// vaultStores: A map of Vault mount configurations.
func Configure(
	ctx *pulumi.Context,
	repositories []*repoConf.Config,
	vaultStores map[string]*vault.Mount,
) ([]string, error) {
	repos := filterRepositories(repositories)

	for name, repository := range repos {
		token, tErr := groupaccesstoken.Create(
			ctx,
			name,
			&groupaccesstoken.CreateOptions{
				Name:        pulumi.String(name),
				Description: pulumi.String(name),
				Group:       repository.AccessPermissions.GitLab.Group,
				Scopes:      repository.AccessPermissions.GitLab.Scopes,
			},
		)
		if tErr != nil {
			log.Err(tErr).
				Msgf("[gitlab][configure] error creating GitLab access token for repository: %s", repository.Name)
			return nil, tErr
		}

		pulumi.All(vaultStores[name].Path, token.Token).
			ApplyT(func(all []any) error {
				path, _ := all[0].(string)
				token, _ := all[1].(string)

				value, _ := json.Marshal(map[string]string{
					"token": token,
				})
				_, _ = vaultSecret.Create(ctx, &vaultSecret.CreateOptions{
					Path:  path,
					Key:   "gitlab",
					Value: pulumi.String(value),
					PulumiOptions: []pulumi.ResourceOption{
						pulumi.Provider(config.VaultProvider),
					},
				})
				return nil
			})
	}

	return slices.Collect(maps.Keys(repos)), nil
}

// filterRepositories filters the given repositories to include only those that we want to create GitLab configurations for.
// repositories: A slice of repository configurations.
func filterRepositories(repositories []*repoConf.Config) map[string]*repoConf.Config {
	repos := make(map[string]*repoConf.Config)
	for _, repository := range repositories {
		repoAccessPermissions := defaults.GetOrDefault(
			repository.AccessPermissions,
			repoConf.AccessPermissionsConfig{},
		)
		if repoAccessPermissions.GitLab != nil && len(repoAccessPermissions.GitLab.Scopes) > 0 {
			repos[repository.Name] = repository
		}
	}

	return repos
}
