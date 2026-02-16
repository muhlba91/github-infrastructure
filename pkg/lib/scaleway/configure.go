package scaleway

import (
	"fmt"

	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	scalewayConf "github.com/muhlba91/github-infrastructure/pkg/model/config/scaleway"
	"github.com/muhlba91/github-infrastructure/pkg/model/scaleway"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scw "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/rs/zerolog/log"
)

// Configure sets up Scaleway resources based on the provided configuration.
// ctx: Pulumi context for resource management.
// repositories: List of repository configurations.
// vaultStores: Map of Vault mount configurations.
// scalewayConfig: Scaleway configuration details.
func Configure(ctx *pulumi.Context,
	repositories []*repoConf.Config,
	vaultStores map[string]*vault.Mount,
	scalewayConfig *scalewayConf.Config,
) (map[string][]string, error) {
	providers := createProviders(ctx, scalewayConfig)

	googleRepositoryProjects := createScalewayRepositoryProjects(repositories, scalewayConfig)

	projects := make(map[string][]string)
	for _, repositoryProject := range googleRepositoryProjects {
		pErr := configureProject(
			ctx,
			repositoryProject,
			vaultStores[*repositoryProject.Repository],
			scalewayConfig,
			providers[*repositoryProject.Name],
		)
		if pErr != nil {
			return nil, pErr
		}

		projectRepositoryMapping, prmOk := projects[*repositoryProject.Name]
		if !prmOk {
			projectRepositoryMapping = []string{}
		}
		projects[*repositoryProject.Name] = append(projectRepositoryMapping, *repositoryProject.Repository)

		for linkedProject := range repositoryProject.LinkedProjects {
			linkedProjectRepositoryMapping, plpOk := projects[linkedProject]
			if !plpOk {
				linkedProjectRepositoryMapping = []string{}
			}
			projects[linkedProject] = append(linkedProjectRepositoryMapping, *repositoryProject.Repository)
		}
	}

	return projects, nil
}

// createProviders initializes GCP providers for each project specified in the configuration.
// ctx: Pulumi context for resource management.
// scalewayConfig: Scaleway configuration details.
func createProviders(ctx *pulumi.Context, scalewayConfig *scalewayConf.Config) map[string]*scw.Provider {
	providers := make(map[string]*scw.Provider)

	for project, pid := range scalewayConfig.Projects {
		provider, _ := scw.NewProvider(ctx, fmt.Sprintf("scaleway-provider-%s", project), &scw.ProviderArgs{
			OrganizationId: pulumi.String(*scalewayConfig.OrganizationID),
			ProjectId:      pulumi.String(*pid),
		})
		providers[project] = provider
	}

	return providers
}

// filterRepositoryByAllowedProjects checks if the repository's Scaleway project
// is included in the list of allowed projects from the configuration.
// repoAccessPermissionsScaleway: Scaleway access configuration for the repository.
// scalewayConfig: Scaleway configuration details.
func filterRepositoryByAllowedProjects(
	repoAccessPermissionsScaleway repoConf.ScalewayAccessConfig,
	scalewayConfig *scalewayConf.Config,
) bool {
	mainProject := repoAccessPermissionsScaleway.Project
	if mainProject == nil || scalewayConfig.Projects[*mainProject] == nil {
		log.Error().Msgf("[scaleway][%v] the repository references an unconfigured project", *mainProject)
		return false
	}

	linkedProjects := repoAccessPermissionsScaleway.LinkedProjects
	if linkedProjects == nil {
		return true
	}

	for project := range linkedProjects {
		if scalewayConfig.Projects[project] == nil {
			log.Error().Msgf("[scaleway][%v] the repository references an unconfigured project", project)
			return false
		}
	}

	return true
}

// createScalewayRepositoryProjects constructs a map of Scaleway repository projects
// based on the provided repository configurations and Scaleway configuration.
// repositories: List of repository configurations.
// scalewayConfig: Scaleway configuration details.
func createScalewayRepositoryProjects(
	repositories []*repoConf.Config,
	scalewayConfig *scalewayConf.Config,
) map[string]*scaleway.RepositoryProject {
	scalewayRepositoryProjects := make(map[string]*scaleway.RepositoryProject)
	for _, repository := range repositories {
		repoAccessPermissions := defaults.GetOrDefault(
			repository.AccessPermissions,
			repoConf.AccessPermissionsConfig{},
		)
		repoAccessPermissionsScaleway := defaults.GetOrDefault(
			repoAccessPermissions.Scaleway,
			repoConf.ScalewayAccessConfig{},
		)

		if repoAccessPermissionsScaleway.Project != nil && *repoAccessPermissionsScaleway.Project != "" &&
			filterRepositoryByAllowedProjects(repoAccessPermissionsScaleway, scalewayConfig) {
			project := defaults.GetOrDefault(repoAccessPermissionsScaleway.Project, "")
			region := defaults.GetOrDefault(
				repoAccessPermissionsScaleway.Region,
				*scalewayConfig.DefaultRegion,
			)
			zone := defaults.GetOrDefault(
				repoAccessPermissionsScaleway.Zone,
				*scalewayConfig.DefaultZone,
			)

			repoLinkedProjects := defaults.GetOrDefault(
				&repoAccessPermissionsScaleway.LinkedProjects,
				map[string]repoConf.ScalewayLinkedAccessConfig{},
			)
			linkedProjects := make(map[string]*scaleway.RepositoryLinkedProject)
			for linkedProject, linkedConfig := range repoLinkedProjects {
				linkedProjects[linkedProject] = &scaleway.RepositoryLinkedProject{
					IAMPermissions: linkedConfig.IAMPermissions,
					AccessLevel:    linkedConfig.AccessLevel,
				}
			}

			scalewayRepositoryProjects[repository.Name] = &scaleway.RepositoryProject{
				Repository:     &repository.Name,
				Name:           &project,
				OrganizationID: scalewayConfig.OrganizationID,
				Region:         &region,
				Zone:           &zone,
				IAMPermissions: append(
					repoAccessPermissionsScaleway.IAMPermissions,
					defaultProjectPermissions...,
				),
				LinkedProjects: linkedProjects,
			}
		}
	}

	return scalewayRepositoryProjects
}
