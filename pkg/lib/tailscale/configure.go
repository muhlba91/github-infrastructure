package tailscale

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	vaultSecret "github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	tsProvider "github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// maxOauthDescriptionLength defines the maximum length for the Tailscale OAuth client description.
const maxOauthDescriptionLength = 50

// Configure sets up Tailscale configurations for the specified repositories.
// ctx: The Pulumi context for resource management.
// repositories: A slice of repository configurations.
// vaultStores: A map of Vault mount configurations.
func Configure(
	ctx *pulumi.Context,
	repositories []*repoConf.Config,
	vaultStores map[string]*vault.Mount,
) ([]*string, error) {
	repos := filterRepositories(repositories)

	for _, repository := range repos {
		oauthClient, oErr := tsProvider.NewOauthClient(
			ctx,
			fmt.Sprintf("tailscale-oauth-client-%s", *repository),
			&tsProvider.OauthClientArgs{
				Description: pulumi.String((*repository)[:min(maxOauthDescriptionLength, len(*repository))]),
				Scopes:      pulumi.ToStringArray([]string{"all"}),
			},
		)
		if oErr != nil {
			log.Err(oErr).
				Msgf("[tailscale][configure] error creating Tailscale OAuth client for repository: %s", *repository)
			return nil, oErr
		}

		pulumi.All(vaultStores[*repository].Path, oauthClient.ID().ToStringOutput(), oauthClient.Key).
			ApplyT(func(all []any) error {
				path, _ := all[0].(string)
				id, _ := all[1].(string)
				key, _ := all[2].(string)

				value, _ := json.Marshal(map[string]string{
					"oauth_client_id": id,
					"oauth_secret":    key,
				})
				_, _ = vaultSecret.Create(ctx, &vaultSecret.CreateOptions{
					Path:  path,
					Key:   "tailscale",
					Value: pulumi.String(value),
					PulumiOptions: []pulumi.ResourceOption{
						pulumi.Provider(config.VaultProvider),
					},
				})
				return nil
			})
	}

	return repos, nil
}

// filterRepositories filters the given repositories to include only those that we want to create Tailscale configurations for.
// repositories: A slice of repository configurations.
func filterRepositories(repositories []*repoConf.Config) []*string {
	var repos []*string
	for _, repository := range repositories {
		repoAccessPermissions := defaults.GetOrDefault(
			repository.AccessPermissions,
			repoConf.AccessPermissionsConfig{},
		)
		if defaults.GetOrDefault(repoAccessPermissions.Tailscale, false) {
			repos = append(repos, &repository.Name)
		}
	}

	return repos
}
