package repository

// RulesetConfig defines repository branch protections config.
type RulesetConfig struct {
	// Enabled indicates whether the ruleset is enabled.
	Enabled bool `yaml:"enabled"`
	// Patterns defines the branch patterns to which the ruleset applies.
	Patterns []string `yaml:"patterns,omitempty"`
	// RestrictCreation indicates whether to restrict creation.
	RestrictCreation *bool `yaml:"restrictCreation,omitempty"`
	// AllowForcePush indicates whether to allow force push.
	AllowForcePush *bool `yaml:"allowForcePush,omitempty"`
	// RequireConversationResolution indicates whether to require conversation resolution.
	RequireConversationResolution *bool `yaml:"requireConversationResolution,omitempty"`
	// RequireSignedCommits indicates whether to require signed commits.
	RequireSignedCommits *bool `yaml:"requireSignedCommits,omitempty"`
	// RequireCodeOwnerReview indicates whether to require code owner review.
	RequireCodeOwnerReview *bool `yaml:"requireCodeOwnerReview,omitempty"`
	// ApprovingReviewCount defines the number of approving reviews required.
	ApprovingReviewCount *int `yaml:"approvingReviewCount,omitempty"`
	// RequireLastPushApproval indicates whether to require last push approval.
	RequireLastPushApproval *bool `yaml:"requireLastPushApproval,omitempty"`
	// RequireUpdatedBranchBeforeMerge indicates whether to require an updated branch before merge.
	RequireUpdatedBranchBeforeMerge *bool `yaml:"requireUpdatedBranchBeforeMerge,omitempty"`
	// EnableMergeQueue indicates whether to enable the merge queue.
	EnableMergeQueue *bool `yaml:"enableMergeQueue,omitempty"`
	// RequiredChecks defines the required status checks.
	RequiredChecks []string `yaml:"requiredChecks,omitempty"`
	// AllowBypass indicates whether to allow bypassing the ruleset.
	AllowBypass *bool `yaml:"allowBypass,omitempty"`
	// AllowBypassIntegrations defines the integrations that are allowed to bypass the ruleset.
	AllowBypassIntegrations []int `yaml:"allowBypassIntegrations,omitempty"`
	// EnableGitstreamIntegration indicates whether to enable the Gitstream integration.
	EnableGitstreamIntegration *bool `yaml:"enableGitstreamIntegration,omitempty"`
	// EnableWipIntegration indicates whether to enable the WIP integration.
	EnableWipIntegration *bool `yaml:"enableWipIntegration,omitempty"`
}
