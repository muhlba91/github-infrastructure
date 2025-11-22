package google

import (
	gcpConf "github.com/muhlba91/github-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// configureProject sets up Google Cloud project resources based on the provided configuration.
// ctx: Pulumi context for resource management.
// project: Google Cloud project details.
// workloadIdentityPool: Workload identity pool for the project.
// vaultStore: Vault mount where secrets will be stored.
// repositoriesConfig: Repository configuration details.
// gcpConfig: Google Cloud configuration details.
// provider: GCP provider for resource creation.
func configureProject(ctx *pulumi.Context,
	project *google.RepositoryProject,
	workloadIdentityPool *google.WorkloadIdentityPool,
	vaultStore *vault.Mount,
	repositoriesConfig *repositories.Config,
	gcpConfig *gcpConf.Config,
	provider *gcp.Provider,
) error {
	serviceAccount, saErr := createProjectIAM(
		ctx,
		project,
		workloadIdentityPool,
		vaultStore,
		repositoriesConfig,
		provider,
	)
	if saErr != nil {
		return saErr
	}

	if defaults.GetOrDefault(gcpConfig.AllowHMACKeys, false) && defaults.GetOrDefault(project.HMACKey, false) {
		hmacErr := createHMACKey(ctx, project, serviceAccount, vaultStore, provider)
		if hmacErr != nil {
			return hmacErr
		}
	}

	return nil
}
