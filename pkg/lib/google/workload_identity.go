package google

import (
	"fmt"
	"maps"
	"strings"

	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ConfigureWorkloadIdentityPools sets up Workload Identity Pools for the given Google Cloud projects.
// ctx: Pulumi context for resource management.
// googleRepositoryProjects: Map of repository projects with their configurations.
// providers: Map of GCP providers configured for specific projects.
// enabledServices: Map of enabled services for each project.
func ConfigureWorkloadIdentityPools(ctx *pulumi.Context,
	googleRepositoryProjects map[string]*google.RepositoryProject,
	providers map[string]*gcp.Provider,
	enabledServices map[string][]pulumi.Resource,
) (map[string]*google.WorkloadIdentityPool, error) {
	workloadIdentities := make(map[string]*google.WorkloadIdentityPool)

	uniqueProjects := make(map[string]bool)
	for _, repositoryProject := range googleRepositoryProjects {
		uniqueProjects[*repositoryProject.Name] = true
	}
	for repositoryProject := range maps.Keys(uniqueProjects) {
		oidc, oErr := createProjectGitHubOidc(
			ctx,
			repositoryProject,
			providers[repositoryProject],
			enabledServices[repositoryProject],
		)
		if oErr != nil {
			return nil, oErr
		}
		workloadIdentities[repositoryProject] = oidc
	}

	return workloadIdentities, nil
}

// createProjectGitHubOidc creates a Workload Identity Pool for GitHub OIDC integration in the specified project.
// ctx: Pulumi context for resource management.
// project: The Google Cloud project ID.
// provider: GCP provider configured for the specific project.
// enabledServices: List of enabled services for the project.
func createProjectGitHubOidc(ctx *pulumi.Context,
	project string,
	provider *gcp.Provider,
	enabledServices []pulumi.Resource,
) (*google.WorkloadIdentityPool, error) {
	poolPostfixRes, ppErr := random.CreateString(
		ctx,
		fmt.Sprintf("random-string-gcp-iam-identity-pool-%s", project),
		&random.StringOptions{
			Length:  postfixLength,
			Special: false,
		},
	)
	if ppErr != nil {
		return nil, ppErr
	}
	poolPostfix, _ := poolPostfixRes.Text.ApplyT(strings.ToLower).(pulumi.StringOutput)

	pool, pErr := iam.NewWorkloadIdentityPool(ctx,
		fmt.Sprintf("gcp-iam-identity-pool-%s", project),
		&iam.WorkloadIdentityPoolArgs{
			WorkloadIdentityPoolId: pulumi.Sprintf(`github-%s`, poolPostfix),
			DisplayName:            pulumi.String("GitHub Identity Pool"),
			Description:            pulumi.String("Workload Identity pool to federate GitHub repositories"),
			Project:                pulumi.StringPtr(project),
		}, pulumi.Provider(provider), pulumi.DependsOn(enabledServices))
	if pErr != nil {
		return nil, pErr
	}

	poolProvider, prErr := iam.NewWorkloadIdentityPoolProvider(ctx,
		fmt.Sprintf("gcp-iam-identity-provider-%s", project),
		&iam.WorkloadIdentityPoolProviderArgs{
			WorkloadIdentityPoolId:         pool.WorkloadIdentityPoolId,
			WorkloadIdentityPoolProviderId: pulumi.Sprintf(`github-actions-%s`, poolPostfix),
			DisplayName:                    pulumi.String("GitHub Actions Provider"),
			Description:                    pulumi.String("Workload Identity Provider to federate GitHub Actions"),
			Oidc: &iam.WorkloadIdentityPoolProviderOidcArgs{
				IssuerUri: pulumi.String("https://token.actions.githubusercontent.com"),
			},
			AttributeMapping: pulumi.ToStringMap(map[string]string{
				"google.subject":             "assertion.sub",
				"attribute.actor":            "assertion.actor",
				"attribute.repository_owner": "assertion.repository_owner",
				"attribute.repository":       "assertion.repository",
			}),
			Project: pulumi.StringPtr(project),
		}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{pool}))
	if prErr != nil {
		return nil, prErr
	}

	return &google.WorkloadIdentityPool{
		WorkloadIdentityPool:     pool,
		WorkloadIdentityProvider: poolProvider,
	}, nil
}
