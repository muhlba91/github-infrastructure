package google

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// createProjectIAM creates IAM roles and service accounts for Continuous Integration in the specified Google Cloud project.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// workloadIdentityPool: Workload Identity Pool for the project.
// vaultStore: Vault mount configuration.
// repositoriesConfig: Configuration for the GitHub repositories.
// provider: GCP provider configured for the specific project.
func createProjectIAM(ctx *pulumi.Context,
	project *google.RepositoryProject,
	workloadIdentityPool *google.WorkloadIdentityPool,
	vaultStore *vault.Mount,
	repositoriesConfig *repositories.Config,
	provider *gcp.Provider,
) (*serviceaccount.Account, error) {
	truncatedRepository := (*project.Repository)[:min(maxRepositoryLength, len(*project.Repository))]

	ciPostfixRes, ciPfErr := random.CreateString(
		ctx,
		fmt.Sprintf("random-string-gcp-iam-role-ci-%s-%s", *project.Repository, *project.Name),
		&random.StringOptions{
			Length:  postfixLength,
			Special: false,
		},
	)
	if ciPfErr != nil {
		log.Err(ciPfErr).
			Msgf("[google][iam] error creating random string for Google Cloud IAM role postfix for project: %s", *project.Name)
		return nil, ciPfErr
	}
	ciPostfix, _ := ciPostfixRes.Text.ApplyT(strings.ToLower).(pulumi.StringOutput)

	gcpProjects := []string{*project.Name}
	for linkedProject := range project.LinkedProjects {
		gcpProjects = append(gcpProjects, linkedProject)
	}

	ciRoles, rErr := createCIRoles(ctx, project, &gcpProjects, &truncatedRepository, ciPostfix, provider)
	if rErr != nil {
		log.Err(rErr).Msgf("[google][iam] error creating Google Cloud IAM roles for project: %s", *project.Name)
		return nil, rErr
	}

	serviceAccount, saErr := createServiceAccount(
		ctx,
		project,
		&gcpProjects,
		&truncatedRepository,
		ciPostfix,
		ciRoles,
		workloadIdentityPool,
		repositoriesConfig,
		provider,
	)
	if saErr != nil {
		log.Err(saErr).
			Msgf("[google][iam] error creating Google Cloud IAM service account for project: %s", *project.Name)
		return nil, saErr
	}

	pulumi.All(vaultStore.Path, workloadIdentityPool.WorkloadIdentityProvider.Name,
		serviceAccount.Email).ApplyT(func(all []any) error {
		path, _ := all[0].(string)
		providerName, _ := all[1].(string)
		email, _ := all[2].(string)

		value, _ := json.Marshal(map[string]string{
			"workload_identity_provider": providerName,
			"ci_service_account":         email,
			"region":                     *project.Region,
		})

		_, err := secret.Create(ctx, &secret.CreateOptions{
			Path:  path,
			Key:   "google-cloud",
			Value: pulumi.String(value),
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(config.VaultProvider),
			},
		})
		return err
	})

	return serviceAccount, nil
}

// createCIRoles creates custom IAM roles for Continuous Integration in the specified Google Cloud projects.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// gcpProjects: List of Google Cloud project IDs to create roles in.
// truncatedRepository: Truncated repository name for role naming.
// ciPostfix: Postfix string for role ID uniqueness.
// provider: GCP provider configured for the specific project.
func createCIRoles(
	ctx *pulumi.Context,
	project *google.RepositoryProject,
	gcpProjects *[]string,
	truncatedRepository *string,
	ciPostfix pulumi.StringOutput,
	provider *gcp.Provider,
) (map[string]*projects.IAMCustomRole, error) {
	ciRoles := make(map[string]*projects.IAMCustomRole)

	for _, projName := range *gcpProjects {
		var permissions []string
		linkedProj, ok := project.LinkedProjects[projName]
		if projName == *project.Name || (ok && linkedProj.AccessLevel == "full") {
			permissions = project.IAMPermissions
		} else {
			permissions = append([]string{}, linkedProj.IAMPermissions...)
			permissions = append(permissions, defaultPermissions...)
		}

		role, roleErr := projects.NewIAMCustomRole(
			ctx,
			fmt.Sprintf("gcp-iam-role-ci-%s-%s", *project.Repository, projName),
			&projects.IAMCustomRoleArgs{
				RoleId: pulumi.Sprintf("ci.%s.%s", strings.ReplaceAll(*truncatedRepository, "-", "_"), ciPostfix),
				Title:  pulumi.String(fmt.Sprintf("GitHub Repository: %s", *project.Repository)),
				Description: pulumi.String(
					fmt.Sprintf("Continuous Integration role for the GitHub repository: %s", *project.Repository),
				),
				Stage:       pulumi.String("GA"),
				Permissions: pulumi.ToStringArray(permissions),
				Project:     pulumi.String(projName),
			},
			pulumi.Provider(provider),
		)
		if roleErr != nil {
			log.Err(roleErr).Msgf("[google][iam] error creating Google Cloud IAM role for project: %s", projName)
			return nil, roleErr
		}

		ciRoles[projName] = role
	}

	return ciRoles, nil
}

// createServiceAccount creates a service account and assigns IAM roles for Continuous Integration.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// gcpProjects: List of Google Cloud project IDs to assign roles in.
// truncatedRepository: Truncated repository name for account naming.
// ciPostfix: Postfix string for account ID uniqueness.
// ciRoles: Map of IAM custom roles created for CI.
// workloadIdentityPool: Workload Identity Pool for the project.
// repositoriesConfig: Configuration for the GitHub repositories.
// provider: GCP provider configured for the specific project.
func createServiceAccount(
	ctx *pulumi.Context,
	project *google.RepositoryProject,
	gcpProjects *[]string,
	truncatedRepository *string,
	ciPostfix pulumi.StringOutput,
	ciRoles map[string]*projects.IAMCustomRole,
	workloadIdentityPool *google.WorkloadIdentityPool,
	repositoriesConfig *repositories.Config,
	provider *gcp.Provider,
) (*serviceaccount.Account, error) {
	serviceAccount, saErr := serviceaccount.NewAccount(
		ctx,
		fmt.Sprintf("gcp-iam-serviceaccount-ci-%s-%s", *project.Repository, *project.Name),
		&serviceaccount.AccountArgs{
			AccountId:   pulumi.Sprintf("ci-%s-%s", *truncatedRepository, ciPostfix),
			DisplayName: pulumi.String(fmt.Sprintf("GitHub Repository: %s", *project.Repository)),
			Description: pulumi.String(
				fmt.Sprintf(
					"Continuous Integration Service Account for the GitHub repository: %s",
					*project.Repository,
				),
			),
			Project: pulumi.String(*project.Name),
		},
		pulumi.Provider(provider),
	)
	if saErr != nil {
		log.Err(saErr).
			Msgf("[google][iam] error creating Google Cloud IAM service account for project: %s", *project.Name)
		return nil, saErr
	}

	for _, projName := range *gcpProjects {
		_, mbrErr := projects.NewIAMMember(
			ctx,
			fmt.Sprintf("gcp-iam-serviceaccount-ci-member-%s-%s", *project.Repository, projName),
			&projects.IAMMemberArgs{
				Project: pulumi.String(projName),
				Role:    ciRoles[projName].ID(),
				Member:  pulumi.Sprintf("serviceAccount:%s", serviceAccount.Email),
			},
			pulumi.Provider(provider),
			pulumi.DependsOn([]pulumi.Resource{
				serviceAccount,
				ciRoles[projName],
			}),
		)
		if mbrErr != nil {
			log.Err(mbrErr).Msgf("[google][iam] error assigning IAM role to service account for project: %s", projName)
			return nil, mbrErr
		}
	}

	_, bindErr := serviceaccount.NewIAMBinding(
		ctx,
		fmt.Sprintf("gcp-iam-identity-member-%s-%s", *project.Repository, *project.Name),
		&serviceaccount.IAMBindingArgs{
			ServiceAccountId: serviceAccount.Name,
			Role:             pulumi.String("roles/iam.workloadIdentityUser"),
			Members: pulumi.StringArray{
				pulumi.Sprintf(
					"principalSet://iam.googleapis.com/%s/attribute.repository/%s/%s",
					workloadIdentityPool.WorkloadIdentityPool.Name,
					*repositoriesConfig.Owner,
					*project.Repository,
				),
			},
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{
			serviceAccount,
			workloadIdentityPool.WorkloadIdentityProvider,
		}),
	)
	if bindErr != nil {
		log.Err(bindErr).
			Msgf("[google][iam] error creating IAM binding for service account in project: %s", *project.Name)
		return nil, bindErr
	}

	return serviceAccount, nil
}
