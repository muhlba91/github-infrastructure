package secrets

import (
	"encoding/json"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	libVault "github.com/muhlba91/github-infrastructure/pkg/lib/vault"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	ghSecret "github.com/muhlba91/pulumi-shared-library/pkg/lib/github/actions/secret"
	vaultSecret "github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nonAlphanumeric matches the characters GitHub doesn't allow in secret names.
var nonAlphanumeric = regexp.MustCompile(`[^A-Za-z0-9]+`)

// Store defines where secrets are written: a Vault mount, the GitHub Actions secrets of a repository, or both.
type Store struct {
	// VaultPath is the path of the Vault KV mount (empty: no Vault).
	VaultPath string
	// VaultMount is the mount if this program creates it (nil: it must exist already).
	VaultMount *vault.Mount
	// Repository is the repository receiving the secrets as GitHub Actions secrets (nil: none).
	Repository *github.Repository
}

// ConfigureStores combines the repositories' Vault mounts and GitHub Actions secrets into stores, keyed like the mounts:
// repositories by name, created additional mounts by libVault.AdditionalMountKey.
// repositories: A slice of repository configurations.
// githubRepositories: A map of GitHub repository resources keyed by repository name.
// mounts: The Vault mounts returned by libVault.ConfigureStores.
func ConfigureStores(
	repositories []*repoConf.Config,
	githubRepositories map[string]*github.Repository,
	mounts map[string]*vault.Mount,
) map[string]*Store {
	stores := make(map[string]*Store)
	for key, mount := range mounts {
		path, additional := libVault.AdditionalMountPath(key)
		if !additional {
			path = libVault.MountPath(key)
		}
		stores[key] = &Store{VaultPath: path, VaultMount: mount}
	}

	for _, repository := range repositories {
		repoAccessPermissions := defaults.GetOrDefault(
			repository.AccessPermissions,
			repoConf.AccessPermissionsConfig{},
		)
		repoGitHubSecrets := defaults.GetOrDefault(repoAccessPermissions.GitHubSecrets, repoConf.GitHubSecretsConfig{})
		if !defaults.GetOrDefault(repoGitHubSecrets.Enabled, false) {
			continue
		}

		store, ok := stores[repository.Name]
		if !ok {
			store = &Store{}
			stores[repository.Name] = store
		}
		store.Repository = githubRepositories[repository.Name]
	}

	return stores
}

// Write writes a secret to the store: to Vault as the key, with the values as JSON fields, and as one GitHub Actions
// secret per value, named <KEY>_<FIELD> in upper case (e.g. google-cloud, region: GOOGLE_CLOUD_REGION).
// A nil store writes nothing.
// ctx: The Pulumi context.
// key: The key of the secret.
// values: The values of the secret, keyed by field.
// dependsOn: Resources the secret depends on.
func (s *Store) Write(ctx *pulumi.Context, key string, values pulumi.StringMap, dependsOn ...pulumi.Resource) error {
	if s == nil {
		return nil
	}

	if s.VaultPath != "" {
		vaultDependsOn := slices.Clone(dependsOn)
		if s.VaultMount != nil {
			vaultDependsOn = append(vaultDependsOn, s.VaultMount)
		}
		value, _ := values.ToStringMapOutput().ApplyT(func(v map[string]string) (string, error) {
			data, err := json.Marshal(v)
			return string(data), err
		}).(pulumi.StringOutput)

		_, err := vaultSecret.Create(ctx, &vaultSecret.CreateOptions{
			Path:  s.VaultPath,
			Key:   key,
			Value: value,
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(config.VaultProvider),
				pulumi.DependsOn(vaultDependsOn),
			},
		})
		if err != nil {
			return err
		}
	}

	if s.Repository != nil {
		for _, field := range slices.Sorted(maps.Keys(values)) {
			ghSecret.Create(ctx, &ghSecret.CreateOptions{
				Key:           strings.ToUpper(nonAlphanumeric.ReplaceAllString(key+"_"+field, "_")),
				Value:         values[field],
				Repository:    s.Repository,
				PulumiOptions: []pulumi.ResourceOption{pulumi.DependsOn(dependsOn)},
			})
		}
	}

	return nil
}
