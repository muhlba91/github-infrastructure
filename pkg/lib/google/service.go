package google

import (
	"fmt"
	"maps"
	"slices"

	googleConf "github.com/muhlba91/github-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// EnableProjectServices enables the specified services for the given Google Cloud projects.
// ctx: Pulumi context for resource management.
// googleRepositoryProjects: Map of repository projects with their configurations.
// gcpConfig: Google Cloud configuration details.
// providers: Map of GCP providers configured for specific projects.
func EnableProjectServices(ctx *pulumi.Context,
	googleRepositoryProjects map[string]*google.RepositoryProject,
	gcpConfig *googleConf.Config,
	providers map[string]*gcp.Provider,
) (map[string][]pulumi.Resource, error) {
	enabledServices := make(map[string][]pulumi.Resource)

	for _, project := range gcpConfig.Projects {
		svcsMap := make(map[string]bool)
		for _, repoProject := range googleRepositoryProjects {
			for _, svc := range projectServices(repoProject, project) {
				svcsMap[svc] = true
			}
		}
		services := slices.Sorted(maps.Keys(svcsMap))
		enabled, psErr := enableForProject(ctx, project, services, providers[project])
		if psErr != nil {
			log.Err(psErr).Msgf("[google][service] error enabling services for Google Cloud project: %s", project)
			return nil, psErr
		}
		enabledServices[project] = enabled
	}

	return enabledServices, nil
}

// projectServices returns the services a repository needs in the given project: its own project and linked projects
// with access level "full" get the repository's services, linked projects with any other access level the default
// services, and unrelated projects none.
// repoProject: The repository project configuration.
// project: The Google Cloud project ID.
func projectServices(repoProject *google.RepositoryProject, project string) []string {
	if *repoProject.Name == project {
		return repoProject.EnabledServices
	}
	linkedProject, linked := repoProject.LinkedProjects[project]
	if !linked {
		return nil
	}
	if linkedProject.AccessLevel == "full" {
		return repoProject.EnabledServices
	}
	return defaultServices
}

// enableForProject enables the specified services for a given Google Cloud project.
// ctx: Pulumi context for resource management.
// project: The Google Cloud project ID.
// services: List of services to enable.
// provider: GCP provider configured for the specific project.
func enableForProject(ctx *pulumi.Context,
	project string,
	services []string,
	provider *gcp.Provider,
) ([]pulumi.Resource, error) {
	var enabledServices []pulumi.Resource

	for _, service := range services {
		svc, err := projects.NewService(ctx,
			fmt.Sprintf("gcp-project-service-%s-%s", project, service),
			&projects.ServiceArgs{
				Project: pulumi.String(project),
				Service: pulumi.String(service),
			}, pulumi.Provider(provider))
		if err != nil {
			log.Err(err).
				Msgf("[google][service] error enabling service %s for Google Cloud project: %s", service, project)
			return nil, err
		}
		enabledServices = append(enabledServices, svc)
	}

	return enabledServices, nil
}
