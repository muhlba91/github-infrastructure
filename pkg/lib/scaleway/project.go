package scaleway

import (
	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"
	scalewayConf "github.com/muhlba91/github-infrastructure/pkg/model/config/scaleway"
	"github.com/muhlba91/github-infrastructure/pkg/model/scaleway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scw "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/rs/zerolog/log"
)

// configureProject sets up Scaleway project resources based on the provided configuration.
// ctx: Pulumi context for resource management.
// project: Scaleway project details.
// store: The repository's secret store.
// scalewayConfig: Scaleway configuration details.
// provider: Scaleway provider for resource creation.
func configureProject(ctx *pulumi.Context,
	project *scaleway.RepositoryProject,
	store *secrets.Store,
	scalewayConfig *scalewayConf.Config,
	provider *scw.Provider,
) error {
	_, saErr := createProjectIAM(
		ctx,
		project,
		store,
		scalewayConfig,
		provider,
	)
	if saErr != nil {
		log.Err(saErr).Msgf("[scaleway][project] error configuring IAM for Scaleway project: %s", *project.Name)
		return saErr
	}

	return nil
}
