package tailscale

import (
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"

	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	tsProvider "github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// maxOauthDescriptionLength defines the maximum length for the Tailscale OAuth client description.
const maxOauthDescriptionLength = 50

// Configure sets up Tailscale configurations for the specified repositories.
// ctx: The Pulumi context for resource management.
// repositories: A slice of repository configurations.
// stores: The secret stores (see secrets.ConfigureStores).
func Configure(
	ctx *pulumi.Context,
	repositories []*repoConf.Config,
	stores map[string]*secrets.Store,
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

		stErr := stores[*repository].Write(ctx, "tailscale", pulumi.StringMap{
			"oauth_client_id": oauthClient.ID().ToStringOutput(),
			"oauth_secret":    oauthClient.Key,
		})
		if stErr != nil {
			log.Err(stErr).
				Msgf("[tailscale][configure] error storing Tailscale OAuth client for repository: %s", *repository)
			return nil, stErr
		}
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
