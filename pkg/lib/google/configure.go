package google

import (
	"fmt"
	"slices"

	googleConf "github.com/muhlba91/github-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// Configure sets up Google Cloud resources based on the provided configuration.
// ctx: Pulumi context for resource management.
// repositories: List of repository configurations.
// vaultStores: Map of Vault mount configurations.
// gcpConfig: Google Cloud configuration details.
// repositoriesConfig: Repository configuration details.
func Configure(ctx *pulumi.Context,
	repositories []*repoConf.Config,
	vaultStores map[string]*vault.Mount,
	gcpConfig *googleConf.Config,
	repositoriesConfig *repositories.Config,
) (map[string][]string, error) {
	providers := createProviders(ctx, gcpConfig)

	googleRepositoryProjects := createGoogleRepositoryProjects(repositories, gcpConfig)

	enabledServices, enableErr := EnableProjectServices(ctx, googleRepositoryProjects, gcpConfig, providers)
	if enableErr != nil {
		return nil, enableErr
	}

	workloadIdentities, wiErr := ConfigureWorkloadIdentityPools(
		ctx,
		googleRepositoryProjects,
		providers,
		enabledServices,
	)
	if wiErr != nil {
		return nil, wiErr
	}

	projects := make(map[string][]string)
	for _, repositoryProject := range googleRepositoryProjects {
		pErr := configureProject(
			ctx,
			repositoryProject,
			workloadIdentities[*repositoryProject.Name],
			vaultStores[*repositoryProject.Repository],
			repositoriesConfig,
			gcpConfig,
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
// gcpConfig: Google Cloud configuration details.
func createProviders(ctx *pulumi.Context, gcpConfig *googleConf.Config) map[string]*gcp.Provider {
	providers := make(map[string]*gcp.Provider)

	for _, project := range gcpConfig.Projects {
		provider, _ := gcp.NewProvider(ctx, fmt.Sprintf("gcp-provider-%s", project), &gcp.ProviderArgs{
			Project: pulumi.String(project),
		})
		providers[project] = provider
	}

	return providers
}

// filterRepositoryByAllowedProjects checks if the repository's Google project
// is included in the list of allowed projects from the configuration.
// repoAccessPermissionsGoogle: Google access configuration for the repository.
// gcpConfig: Google Cloud configuration details.
func filterRepositoryByAllowedProjects(
	repoAccessPermissionsGoogle repoConf.GoogleAccessConfig,
	gcpConfig *googleConf.Config,
) bool {
	mainProject := repoAccessPermissionsGoogle.Project
	if mainProject == nil || !slices.Contains(gcpConfig.Projects, *mainProject) {
		log.Error().Msgf("[google][%v] the repository references an unconfigured project", *mainProject)
		return false
	}

	linkedProjects := repoAccessPermissionsGoogle.LinkedProjects
	if linkedProjects == nil {
		return true
	}

	for project := range linkedProjects {
		if !slices.Contains(gcpConfig.Projects, project) {
			log.Error().Msgf("[google][%v] the repository references an unconfigured project", project)
			return false
		}
	}

	return true
}

// createGoogleRepositoryProjects constructs a map of Google repository projects
// based on the provided repository configurations and GCP configuration.
// repositories: List of repository configurations.
// gcpConfig: Google Cloud configuration details.
func createGoogleRepositoryProjects(
	repositories []*repoConf.Config,
	gcpConfig *googleConf.Config,
) map[string]*google.RepositoryProject {
	googleRepositoryProjects := make(map[string]*google.RepositoryProject)
	for _, repository := range repositories {
		repoAccessPermissions := defaults.GetOrDefault(
			repository.AccessPermissions,
			repoConf.AccessPermissionsConfig{},
		)
		repoAccessPermissionsGoogle := defaults.GetOrDefault(
			repoAccessPermissions.Google,
			repoConf.GoogleAccessConfig{},
		)

		if repoAccessPermissionsGoogle.Project != nil && *repoAccessPermissionsGoogle.Project != "" &&
			filterRepositoryByAllowedProjects(repoAccessPermissionsGoogle, gcpConfig) {
			project := defaults.GetOrDefault(repoAccessPermissionsGoogle.Project, "")
			region := defaults.GetOrDefault(
				repoAccessPermissionsGoogle.Region,
				*gcpConfig.DefaultRegion,
			)

			repoLinkedProjects := defaults.GetOrDefault(
				&repoAccessPermissionsGoogle.LinkedProjects,
				map[string]repoConf.GoogleLinkedAccessConfig{},
			)
			linkedProjects := make(map[string]*google.RepositoryLinkedProject)
			for linkedProject, linkedConfig := range repoLinkedProjects {
				linkedProjects[linkedProject] = &google.RepositoryLinkedProject{
					IAMPermissions: linkedConfig.IAMPermissions,
					AccessLevel:    linkedConfig.AccessLevel,
				}
			}

			googleRepositoryProjects[repository.Name] = &google.RepositoryProject{
				Repository: &repository.Name,
				Name:       &project,
				Region:     &region,
				IAMPermissions: append(
					repoAccessPermissionsGoogle.IAMPermissions,
					defaultPermissions...,
				),
				EnabledServices: append(
					repoAccessPermissionsGoogle.EnabledServices,
					defaultServices...,
				),
				LinkedProjects: linkedProjects,
				HMACKey:        repoAccessPermissionsGoogle.HMACKey,
			}
		}
	}

	return googleRepositoryProjects
}
