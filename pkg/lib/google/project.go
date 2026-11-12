package google

import (
	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"
	gcpConf "github.com/muhlba91/github-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// configureProject sets up Google Cloud project resources based on the provided configuration.
// ctx: Pulumi context for resource management.
// project: Google Cloud project details.
// workloadIdentityPool: Workload identity pool for the project.
// store: The repository's secret store.
// repositoriesConfig: Repository configuration details.
// gcpConfig: Google Cloud configuration details.
// provider: GCP provider for resource creation.
func configureProject(ctx *pulumi.Context,
	project *google.RepositoryProject,
	workloadIdentityPool *google.WorkloadIdentityPool,
	store *secrets.Store,
	repositoriesConfig *repositories.Config,
	gcpConfig *gcpConf.Config,
	provider *gcp.Provider,
) error {
	serviceAccount, saErr := createProjectIAM(
		ctx,
		project,
		workloadIdentityPool,
		store,
		repositoriesConfig,
		provider,
	)
	if saErr != nil {
		log.Err(saErr).Msgf("[google][project] error configuring IAM for Google Cloud project: %s", *project.Name)
		return saErr
	}

	if defaults.GetOrDefault(gcpConfig.AllowHMACKeys, false) && defaults.GetOrDefault(project.HMACKey, false) {
		hmacErr := createHMACKey(ctx, project, serviceAccount, store, provider)
		if hmacErr != nil {
			log.Err(hmacErr).
				Msgf("[google][project] error creating HMAC key for service account in project: %s", *project.Name)
			return hmacErr
		}
	}

	return nil
}
