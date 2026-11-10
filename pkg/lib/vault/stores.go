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
// It returns the repositories' mounts keyed by the repository name and the created additional mounts keyed by
// AdditionalMountKey.
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

		repositoryMounts := make(map[string]*vault.Mount)
		for path := range additionalMounts {
			additionalMount, addErr := store.Create(ctx, path, &store.CreateOptions{
				Path:        pulumi.String(path),
				Description: pulumi.String("Secrets for: " + path),
				PulumiOptions: []pulumi.ResourceOption{
					pulumi.Provider(config.VaultProvider),
				},
			})
			if addErr != nil {
				log.Err(addErr).Msgf("[vault][store] error creating vault store for additional mount: %s", path)
				return nil
			}
			repositoryMounts[AdditionalMountKey(path)] = additionalMount
		}

		for _, repository := range repos {
			mount, stErr := store.Create(ctx, repository.Name, &store.CreateOptions{
				Path: pulumi.String(fmt.Sprintf("github-%s", repository.Name)),
				Description: pulumi.String(
					fmt.Sprintf("GitHub repository: %s/%s", *repositoriesConfig.Owner, repository.Name),
				),
				PulumiOptions: []pulumi.ResourceOption{
					pulumi.Provider(config.VaultProvider),
				},
			})
			if stErr != nil {
				log.Err(stErr).Msgf("[vault][store] error creating vault store for repository: %s", repository.Name)
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
				log.Err(err).Msgf("[vault][store] error creating vault authentication")
				return nil
			}

			repositoryMounts[repository.Name] = mount
		}

		return repositoryMounts
	})
}

// AdditionalMountKey returns the key of a created additional mount in the map returned by ConfigureStores (the
// repositories' mounts are keyed by the repository name, which never contains a colon).
// path: The mount path.
func AdditionalMountKey(path string) string {
	return "mount:" + path
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
