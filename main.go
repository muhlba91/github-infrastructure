package main

import (
	"errors"
	"maps"
	"slices"

	"github.com/muhlba91/github-infrastructure/pkg/lib/aws"
	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	ghRepos "github.com/muhlba91/github-infrastructure/pkg/lib/github/repositories"
	"github.com/muhlba91/github-infrastructure/pkg/lib/gitlab"
	"github.com/muhlba91/github-infrastructure/pkg/lib/google"
	"github.com/muhlba91/github-infrastructure/pkg/lib/scaleway"
	"github.com/muhlba91/github-infrastructure/pkg/lib/tailscale"
	"github.com/muhlba91/github-infrastructure/pkg/lib/vault"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	vaultProvider "github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// main is the entry point of the Pulumi program.
func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		repositoriesConfig, awsConfig, gcpConfig, scalewayConfig, vaultConfig, repos, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		// repositories
		githubRepositories, ghErr := ghRepos.Create(ctx, repos, repositoriesConfig)
		if ghErr != nil {
			return ghErr
		}

		// vault stores
		vaultStores := vault.ConfigureStores(ctx, repos, githubRepositories, repositoriesConfig, vaultConfig)
		vaultStores.ApplyT(func(stores any) error {
			if stores == nil {
				return errors.New("failed to configure vault stores")
			}
			return nil
		})

		// gitlab access
		gitlabs := vaultStores.ApplyT(func(stores map[string]*vaultProvider.Mount) []string {
			gl, _ := gitlab.Configure(ctx, repos, stores)
			return gl
		})

		// tailscale access
		tailscales := vaultStores.ApplyT(func(stores map[string]*vaultProvider.Mount) []*string {
			ts, _ := tailscale.Configure(ctx, repos, stores)
			return ts
		})

		// google cloud
		googleAllowedProjects := gcpConfig.Projects
		slices.Sort(googleAllowedProjects)
		googleProjects := vaultStores.ApplyT(func(stores map[string]*vaultProvider.Mount) map[string][]string {
			projects, _ := google.Configure(ctx, repos, stores, gcpConfig, repositoriesConfig)
			return projects
		})

		// aws accounts
		awsAllowedAccounts := slices.Collect(maps.Keys(awsConfig.Account))
		slices.Sort(awsAllowedAccounts)
		awsAccounts := vaultStores.ApplyT(func(stores map[string]*vaultProvider.Mount) map[string][]string {
			accounts, _ := aws.Configure(ctx, repos, stores, awsConfig, repositoriesConfig)
			return accounts
		})

		// scaleway projects
		scalewayAllowedProjects := slices.Collect(maps.Keys(scalewayConfig.Projects))
		slices.Sort(scalewayAllowedProjects)
		scalewayProjects := vaultStores.ApplyT(func(stores map[string]*vaultProvider.Mount) map[string][]string {
			projects, _ := scaleway.Configure(ctx, repos, stores, scalewayConfig)
			return projects
		})

		// outputs
		ctx.Export("gitlab", pulumi.ToMap(map[string]any{
			"tokens": gitlabs,
		}))
		ctx.Export("tailscale", pulumi.ToMap(map[string]any{
			"clients": tailscales,
		}))
		ctx.Export("google", pulumi.ToMap(map[string]any{
			"allowed":    googleAllowedProjects,
			"configured": googleProjects,
		}))
		ctx.Export("scaleway", pulumi.ToMap(map[string]any{
			"allowed":    scalewayAllowedProjects,
			"configured": scalewayProjects,
		}))
		ctx.Export("aws", pulumi.ToMap(map[string]any{
			"allowed":    awsAllowedAccounts,
			"configured": awsAccounts,
		}))
		ctx.Export("vault", pulumi.ToMap(map[string]any{
			"projects": vaultStores.ApplyT(func(stores map[string]*vaultProvider.Mount) []string {
				return slices.Collect(maps.Keys(stores))
			}),
		}))
		exportRepositories(ctx, repos)

		return nil
	})
}

// exportRepositories exports a summary of the configured repositories and their access permissions.
// ctx: The Pulumi context used for exporting outputs.
// repos: A slice of repository configurations to be summarized.
func exportRepositories(ctx *pulumi.Context, repos []*repository.Config) {
	repositories := make(map[string]map[string]any)
	for _, repo := range repos {
		repositories[repo.Name] = make(map[string]any)
		repositories[repo.Name]["gitlab"] = repo.AccessPermissions != nil &&
			repo.AccessPermissions.GitLab != nil &&
			len(repo.AccessPermissions.GitLab.Scopes) > 0
		repositories[repo.Name]["google"] = repo.AccessPermissions != nil &&
			repo.AccessPermissions.Google != nil &&
			repo.AccessPermissions.Google.Project != nil
		repositories[repo.Name]["gcs"] = repo.AccessPermissions != nil &&
			repo.AccessPermissions.Google != nil &&
			defaults.GetOrDefault(repo.AccessPermissions.Google.HMACKey, false)
		repositories[repo.Name]["aws"] = repo.AccessPermissions != nil &&
			repo.AccessPermissions.Aws != nil &&
			repo.AccessPermissions.Aws.Account != nil
		repositories[repo.Name]["vault"] = defaults.GetOrDefault(repo.ManageLifecycle, true) &&
			repo.AccessPermissions != nil &&
			repo.AccessPermissions.Vault != nil &&
			defaults.GetOrDefault(repo.AccessPermissions.Vault.Enabled, true)
		repositories[repo.Name]["tailscale"] = repo.AccessPermissions != nil &&
			defaults.GetOrDefault(repo.AccessPermissions.Tailscale, false)
	}

	ctx.Export("repositories", pulumi.ToMapMap(repositories))
}
