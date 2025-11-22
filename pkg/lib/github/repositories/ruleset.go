package repositories

import (
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	libRuleset "github.com/muhlba91/pulumi-shared-library/pkg/lib/github/ruleset"
)

// createRuleset creates a branch ruleset for the given repository based on the provided configuration.
// ctx: The Pulumi context for resource creation.
// name: The name of the ruleset.
// repository: The configuration for the repository.
// repo: The Pulumi GitHub repository resource.
func createRuleset(
	ctx *pulumi.Context,
	name string,
	repository *repository.Config,
	repo *github.Repository,
) error {
	delOnDestroy := true
	patterns := append([]string{libRuleset.DefaultBranchRulesetPattern}, repository.Rulesets.Branch.Patterns...)

	_, err := libRuleset.Create(ctx, name, &libRuleset.CreateOptions{
		Repository:               repo,
		Patterns:                 patterns,
		RestrictCreation:         repository.Rulesets.Branch.RestrictCreation,
		AllowForcePush:           repository.Rulesets.Branch.AllowForcePush,
		SignedCommits:            repository.Rulesets.Branch.RequireSignedCommits,
		CodeOwnerReview:          repository.Rulesets.Branch.RequireCodeOwnerReview,
		ConversationResolution:   repository.Rulesets.Branch.RequireConversationResolution,
		LastPushApproval:         repository.Rulesets.Branch.RequireLastPushApproval,
		ReviewerCount:            repository.Rulesets.Branch.ApprovingReviewCount,
		EnableMergeQueue:         repository.Rulesets.Branch.EnableMergeQueue,
		AllowBypass:              repository.Rulesets.Branch.AllowBypass,
		AllowBypassIntegrations:  repository.Rulesets.Branch.AllowBypassIntegrations,
		UpdatedBranchBeforeMerge: repository.Rulesets.Branch.RequireUpdatedBranchBeforeMerge,
		RequiredChecks:           repository.Rulesets.Branch.RequiredChecks,
		WIPIntegration:           repository.Rulesets.Branch.EnableWipIntegration,
		DeleteOnDestroy:          &delOnDestroy,
	})
	return err
}
