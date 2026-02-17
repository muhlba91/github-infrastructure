package repositories

import (
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	ghUtils "github.com/muhlba91/github-infrastructure/pkg/util/github"
	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"

	libRepo "github.com/muhlba91/pulumi-shared-library/pkg/lib/github/repository"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
)

// defaultVisibility is the default visibility for GitHub repositories.
const defaultVisibility = "public"

// Create creates multiple GitHub repositories based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// repositories: A slice of repository configurations to create.
// repositoriesConfig: General configuration for repositories.
func Create(
	ctx *pulumi.Context,
	repositories []*repository.Config,
	repositoriesConfig *repositories.Config,
) (map[string]*github.Repository, error) {
	repos := make(map[string]*github.Repository)

	for _, repo := range repositories {
		ghRepo, err := create(ctx, repo, repositoriesConfig)
		if err != nil {
			log.Err(err).Msgf("[github][repository] error creating GitHub repository: %s", repo.Name)
			return nil, err
		}
		repos[repo.Name] = ghRepo
	}

	return repos, nil
}

// create creates a single GitHub repository based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// repository: The configuration for the repository to create.
// repositoriesConfig: General configuration for repositories.
func create(
	ctx *pulumi.Context,
	repository *repository.Config,
	repositoriesConfig *repositories.Config,
) (*github.Repository, error) {
	manageLifecycle := defaults.GetOrDefault(repository.ManageLifecycle, true)

	owner := repositoriesConfig.Owner
	resourceName := fmt.Sprintf("%s-%s", *owner, repository.Name)

	if !manageLifecycle && !config.IgnoreUnmanagedRepositories {
		config.Stack.GetOutput(pulumi.String("repositories")).ApplyT(func(repos any) error {
			repoMap, _ := repos.(map[string]any)
			if _, exists := repoMap[repository.Name]; !exists {
				return fmt.Errorf(
					"[ERROR] repository '%s' is not imported yet! Please import it using the following command and re-run Pulumi with IGNORE_UNMANAGED_REPOSITORIES=\"true\": pulumi import github:index/repository:Repository %s %s",
					repository.Name,
					resourceName,
					repository.Name,
				)
			}
			return nil
		})
	}

	defVis := defaultVisibility
	repo, err := libRepo.Create(ctx, resourceName, &libRepo.CreateOptions{
		Name:                    pulumi.String(repository.Name),
		Description:             pulumi.String(repository.Description),
		EnableDiscussions:       pulumi.BoolPtr(defaults.GetOrDefault(repository.EnableDiscussions, false)),
		EnableWiki:              pulumi.BoolPtr(defaults.GetOrDefault(repository.EnableWiki, false)),
		Homepage:                pulumi.StringPtr(defaults.GetOrDefault(repository.Homepage, "")),
		Topics:                  repository.Topics,
		GitHubPagesBranch:       repository.PagesBranch,
		Visibility:              defaults.GetOrDefault(&repository.Visibility, &defVis),
		Protected:               defaults.GetOrDefault(repository.Protected, false),
		AllowRepositoryDeletion: manageLifecycle || !config.AllowRepositoryDeletion,
	})
	if err != nil {
		log.Err(err).
			Msgf("[github][repository] error creating GitHub repository resource for repository: %s", repository.Name)
		return nil, err
	}

	if (ghUtils.HasSubscription(repositoriesConfig.Subscription) || !ghUtils.IsPrivate(repository.Visibility)) &&
		repository.Rulesets.Branch.Enabled {
		rsErr := createRuleset(ctx, fmt.Sprintf("branch-%s-%s", *owner, repository.Name), repository, repo)
		if rsErr != nil {
			log.Err(rsErr).
				Msgf("[github][repository] error creating branch ruleset for repository: %s", repository.Name)
			return nil, rsErr
		}
	}

	return repo, nil
}
