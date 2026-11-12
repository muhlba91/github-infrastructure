package google

import (
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// secretManagerService is the Secret Manager API, enabled in all projects (defaultServices).
const secretManagerService = "secretmanager.googleapis.com"

// secretManagerSeparator separates the repository from the name in the secrets' names: <repository>--<NAME>.
const secretManagerSeparator = "--"

// createSecretManagerAccess grants the CI service account read access to the repository's secrets in the project's
// Secret Manager: the secrets named <repository>--<NAME> (an IAM condition on the secrets' resource names).
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// serviceAccount: The CI service account.
// provider: GCP provider configured for the specific project.
func createSecretManagerAccess(ctx *pulumi.Context,
	project *google.RepositoryProject,
	serviceAccount *serviceaccount.Account,
	provider *gcp.Provider,
) error {
	// the resource names of secrets carry the project number, not its ID
	gcpProject, err := organizations.LookupProject(ctx, &organizations.LookupProjectArgs{
		ProjectId: project.Name,
	}, pulumi.Provider(provider))
	if err != nil {
		log.Err(err).Msgf("[google][secretmanager] error looking up Google Cloud project: %s", *project.Name)
		return err
	}

	_, err = projects.NewIAMMember(
		ctx,
		fmt.Sprintf("gcp-iam-serviceaccount-ci-secretmanager-%s-%s", *project.Repository, *project.Name),
		&projects.IAMMemberArgs{
			Project: pulumi.String(*project.Name),
			Role:    pulumi.String("roles/secretmanager.secretAccessor"),
			Member:  pulumi.Sprintf("serviceAccount:%s", serviceAccount.Email),
			Condition: &projects.IAMMemberConditionArgs{
				Title: pulumi.String(fmt.Sprintf("secrets of %s", *project.Repository)),
				Expression: pulumi.String(fmt.Sprintf(
					`resource.name.startsWith("projects/%s/secrets/%s%s")`,
					gcpProject.Number,
					*project.Repository,
					secretManagerSeparator,
				)),
			},
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{serviceAccount}),
	)
	if err != nil {
		log.Err(err).
			Msgf("[google][secretmanager] error granting Secret Manager access in project: %s", *project.Name)
	}
	return err
}
