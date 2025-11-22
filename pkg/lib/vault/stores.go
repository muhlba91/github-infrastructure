package vault

import (
	"fmt"
	"iter"
	"maps"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	vaultConf "github.com/muhlba91/github-infrastructure/pkg/model/config/vault"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/store"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// ConfigureStores configures Vault secret stores for the given GitHub repositories.
// ctx: The Pulumi context.
// repositories: A slice of repository configurations.
// githubRepositories: A map of GitHub repository resources keyed by repository name.
// repositoriesConfig: The overall repositories configuration.
// vaultConfig: The Vault configuration.
func ConfigureStores(
	ctx *pulumi.Context,
	repositories []*repoConf.Config,
	githubRepositories map[string]*github.Repository,
	repositoriesConfig *repositories.Config,
	vaultConfig *vaultConf.Config,
) pulumi.Output {
	return config.HasVaultConnection.ApplyT(func(hasVaultConn bool) map[string]*vault.Mount {
		if !hasVaultConn {
			return map[string]*vault.Mount{}
		}

		repos, additionalMounts := filterRepositories(repositories)

		for path := range additionalMounts {
			_, addErr := store.Create(ctx, path, &store.CreateArgs{
				Path:        pulumi.String(path),
				Description: pulumi.String("Secrets for: " + path),
				PulumiOptions: []pulumi.ResourceOption{
					pulumi.Provider(config.VaultProvider),
				},
			})
			if addErr != nil {
				log.Err(addErr).Msgf("error creating vault store for additional mount: %s", path)
				return nil
			}
		}

		repositoryMounts := make(map[string]*vault.Mount)
		for _, repository := range repos {
			mount, stErr := store.Create(ctx, repository.Name, &store.CreateArgs{
				Path: pulumi.String(fmt.Sprintf("github-%s", repository.Name)),
				Description: pulumi.String(
					fmt.Sprintf("GitHub repository: %s/%s", *repositoriesConfig.Owner, repository.Name),
				),
				PulumiOptions: []pulumi.ResourceOption{
					pulumi.Provider(config.VaultProvider),
				},
			})
			if stErr != nil {
				log.Err(stErr).Msgf("error creating vault store for repository: %s", repository.Name)
				return nil
			}

			_, err := createAuth(
				ctx,
				repository,
				mount,
				githubRepositories[repository.Name],
				repositoriesConfig,
				vaultConfig,
			)
			if err != nil {
				log.Err(err).Msgf("error creating vault authentication")
				return nil
			}

			repositoryMounts[repository.Name] = mount
		}

		return repositoryMounts
	})
}

// filterRepositories filters the given repositories to include only those that we want to manage the lifecycle for.
// repositories: A slice of repository configurations.
func filterRepositories(repositories []*repoConf.Config) ([]*repoConf.Config, iter.Seq[string]) {
	var repos []*repoConf.Config
	addMountsTmp := make(map[string]bool)
	for _, repository := range repositories {
		repoAccessPermissions := defaults.GetOrDefault(
			repository.AccessPermissions,
			repoConf.AccessPermissionsConfig{},
		)
		repoVaultAccessPermissions := defaults.GetOrDefault(
			repoAccessPermissions.Vault,
			repoConf.VaultAccessPermissionsConfig{},
		)
		vEnabled := defaults.GetOrDefault(repoVaultAccessPermissions.Enabled, true)
		if defaults.GetOrDefault(repository.ManageLifecycle, true) && vEnabled {
			repos = append(repos, repository)

			if repoVaultAccessPermissions.AdditionalMounts != nil {
				for _, mount := range repoVaultAccessPermissions.AdditionalMounts {
					if defaults.GetOrDefault(mount.Create, false) {
						addMountsTmp[mount.Path] = true
					}
				}
			}
		}
	}

	return repos, maps.Keys(addMountsTmp)
}
