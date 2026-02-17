package scaleway

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	scalewayConf "github.com/muhlba91/github-infrastructure/pkg/model/config/scaleway"
	scalewayModel "github.com/muhlba91/github-infrastructure/pkg/model/scaleway"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/scaleway/iam/policy"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	scwmodel "github.com/muhlba91/pulumi-shared-library/pkg/model/scaleway/iam/application"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/scaleway/iam/application"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scw "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/iam"
	"github.com/rs/zerolog/log"
)

// createProjectIAM creates IAM roles and service accounts for Continuous Integration in the specified Scaleway project.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// vaultStore: Vault mount configuration.
// scalewayConfig: Scaleway configuration details.
// provider: Scaleway provider configured for the specific project.
func createProjectIAM(ctx *pulumi.Context,
	project *scalewayModel.RepositoryProject,
	vaultStore *vault.Mount,
	scalewayConfig *scalewayConf.Config,
	provider *scw.Provider,
) (*scwmodel.Application, error) {
	scalewayProjects := []string{*project.Name}
	for linkedProject := range project.LinkedProjects {
		scalewayProjects = append(scalewayProjects, linkedProject)
	}

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

	rErr := createCIPolicies(
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

	pulumi.All(vaultStore.Path, application.Key.AccessKey, application.Key.SecretKey).ApplyT(func(all []any) error {
		path, _ := all[0].(string)
		accessKey, _ := all[1].(string)
		secretKey, _ := all[2].(string)

		value, _ := json.Marshal(map[string]string{
			"access_key":      accessKey,
			"secret_key":      secretKey,
			"region":          *project.Region,
			"zone":            *project.Zone,
			"organization_id": *project.OrganizationID,
			"project_id":      *scalewayConfig.Projects[*project.Name],
		})

		_, err := secret.Create(ctx, &secret.CreateOptions{
			Path:  path,
			Key:   "scaleway",
			Value: pulumi.String(value),
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(config.VaultProvider),
			},
		})
		return err
	})

	return application, nil
}

// createCIPolicies creates custom IAM policies for Continuous Integration in the specified Scaleway projects.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// applicationId: The ID of the application for which the policies are being created.
// scalewayProjects: List of Scaleway project IDs to create policies in.
// scalewayConfig: Scaleway configuration details.
// provider: Scaleway provider configured for the specific project.
func createCIPolicies(
	ctx *pulumi.Context,
	project *scalewayModel.RepositoryProject,
	applicationID pulumi.StringOutput,
	scalewayProjects *[]string,
	scalewayConfig *scalewayConf.Config,
	provider *scw.Provider,
) error {
	for _, projName := range *scalewayProjects {
		var permissions []string
		linkedProj, ok := project.LinkedProjects[projName]
		if projName == *project.Name || (ok && linkedProj.AccessLevel == "full") {
			permissions = project.IAMPermissions
		} else {
			permissions = append([]string{}, linkedProj.IAMPermissions...)
			permissions = append(permissions, defaultProjectPermissions...)
		}

		name := fmt.Sprintf("ci-%s-%s", *project.Repository, projName)
		_, polErr := policy.Create(
			ctx,
			name,
			&policy.CreateOptions{
				Name: pulumi.Sprintf("scw-iam-policy-%s", name),
				Description: pulumi.Sprintf(
					"Continuous Integration policy for the GitHub repository: %s in project: %s",
					*project.Repository,
					projName,
				),
				Rules: []iam.PolicyRuleInput{
					&iam.PolicyRuleArgs{
						OrganizationId:     pulumi.String(*project.OrganizationID),
						PermissionSetNames: pulumi.ToStringArray(defaultOrganizationPermissions),
					},
					&iam.PolicyRuleArgs{
						ProjectIds: pulumi.StringArray{
							pulumi.String(*scalewayConfig.Projects[projName]),
						},
						PermissionSetNames: pulumi.ToStringArray(permissions),
					},
				},
				ApplicationID: applicationID,
				PulumiOptions: []pulumi.ResourceOption{
					pulumi.Provider(provider),
				},
			},
		)
		if polErr != nil {
			log.Err(polErr).Msgf("[scaleway][iam] error creating IAM policy for Scaleway project: %s", projName)
			return polErr
		}
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
	return application.CreateApplication(
		ctx,
		&application.CreateOptions{
			Name:             fmt.Sprintf("scw-iam-application-ci-%s-%s", *project.Repository, *project.Name),
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
		},
	)
}
