package scaleway

import (
	"fmt"
	"slices"

	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"

	scalewayConf "github.com/muhlba91/github-infrastructure/pkg/model/config/scaleway"
	scalewayModel "github.com/muhlba91/github-infrastructure/pkg/model/scaleway"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/scaleway/iam/policy"
	scwmodel "github.com/muhlba91/pulumi-shared-library/pkg/model/scaleway/iam/application"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/scaleway/iam/application"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scw "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/iam"
	"github.com/rs/zerolog/log"
)

// createProjectIAM creates IAM roles and service accounts for Continuous Integration in the specified Scaleway project.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// store: The repository's secret store.
// scalewayConfig: Scaleway configuration details.
// provider: Scaleway provider configured for the specific project.
func createProjectIAM(ctx *pulumi.Context,
	project *scalewayModel.RepositoryProject,
	store *secrets.Store,
	scalewayConfig *scalewayConf.Config,
	provider *scw.Provider,
) (*scwmodel.Application, error) {
	scalewayProjects := []string{*project.Name}
	for linkedProject := range project.LinkedProjects {
		scalewayProjects = append(scalewayProjects, linkedProject)
	}
	slices.Sort(scalewayProjects)

	application, saErr := createApplication(
		ctx,
		project,
		scalewayConfig,
		provider,
	)
	if saErr != nil {
		log.Err(saErr).Msgf("[scaleway][iam] error creating application for Scaleway project: %s", *project.Name)
		return nil, saErr
	}

	rErr := createCIPolicy(
		ctx,
		project,
		application.Application.ID().ToStringOutput(),
		&scalewayProjects,
		scalewayConfig,
		provider,
	)
	if rErr != nil {
		log.Err(rErr).Msgf("[scaleway][iam] error creating IAM policies for Scaleway project: %s", *project.Name)
		return nil, rErr
	}

	stErr := store.Write(ctx, "scaleway", pulumi.StringMap{
		"access_key":      application.Key.AccessKey,
		"secret_key":      application.Key.SecretKey,
		"region":          pulumi.String(*project.Region),
		"zone":            pulumi.String(*project.Zone),
		"organization_id": pulumi.String(*project.OrganizationID),
		"project_id":      pulumi.String(*scalewayConfig.Projects[*project.Name]),
	})
	if stErr != nil {
		log.Err(stErr).Msgf("[scaleway][iam] error storing Scaleway secrets for project: %s", *project.Name)
		return nil, stErr
	}

	return application, nil
}

// createCIPolicy creates a custom IAM policy for Continuous Integration in the specified Scaleway projects.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// applicationId: The ID of the application for which the policy is being created.
// scalewayProjects: List of Scaleway project IDs to create policies in.
// scalewayConfig: Scaleway configuration details.
// provider: Scaleway provider configured for the specific project.
func createCIPolicy(
	ctx *pulumi.Context,
	project *scalewayModel.RepositoryProject,
	applicationID pulumi.StringOutput,
	scalewayProjects *[]string,
	scalewayConfig *scalewayConf.Config,
	provider *scw.Provider,
) error {
	rules := []iam.PolicyRuleInput{
		&iam.PolicyRuleArgs{
			OrganizationId:     pulumi.String(*project.OrganizationID),
			PermissionSetNames: pulumi.ToStringArray(defaultOrganizationPermissions),
		},
	}

	for _, projName := range *scalewayProjects {
		var permissions []string
		linkedProj, ok := project.LinkedProjects[projName]
		if projName == *project.Name || (ok && linkedProj.AccessLevel == "full") {
			permissions = project.IAMPermissions
		} else {
			permissions = append([]string{}, linkedProj.IAMPermissions...)
			permissions = append(permissions, defaultProjectPermissions...)
		}

		rules = append(rules, &iam.PolicyRuleArgs{
			ProjectIds: pulumi.StringArray{
				pulumi.String(*scalewayConfig.Projects[projName]),
			},
			PermissionSetNames: pulumi.ToStringArray(permissions),
		})
	}

	name := fmt.Sprintf("ci-%s", *project.Repository)
	_, polErr := policy.Create(
		ctx,
		name,
		&policy.CreateOptions{
			Name: pulumi.Sprintf("scw-iam-policy-%s", name),
			Description: pulumi.Sprintf(
				"Continuous Integration policy for the GitHub repository: %s",
				*project.Repository,
			),
			Rules:         rules,
			ApplicationID: applicationID,
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(provider),
			},
		},
	)
	if polErr != nil {
		log.Err(polErr).Msgf("[scaleway][iam] error creating IAM policy for repository: %s", *project.Repository)
		return polErr
	}

	return nil
}

// createApplication creates an application and assigns IAM roles for Continuous Integration.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// scalewayConfig: Scaleway configuration details.
// provider: Scaleway provider configured for the specific project.
func createApplication(
	ctx *pulumi.Context,
	project *scalewayModel.RepositoryProject,
	scalewayConfig *scalewayConf.Config,
	provider *scw.Provider,
) (*scwmodel.Application, error) {
	name := fmt.Sprintf("ci-%s-%s", *project.Repository, *project.Name)
	if len(name) > maxApplicationNameLength {
		name = name[:maxApplicationNameLength]
	}

	return application.CreateApplication(
		ctx,
		&application.CreateOptions{
			Name:             name,
			DefaultProjectID: pulumi.String(*scalewayConfig.Projects[*project.Name]),
			Description: pulumi.String(
				fmt.Sprintf(
					"Continuous Integration Application for the GitHub repository: %s",
					*project.Repository,
				),
			),
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(provider),
			},
			Labels: append(commonLabels(), fmt.Sprintf("repository=%s", *project.Repository)),
		},
	)
}
