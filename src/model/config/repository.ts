/**
 * Defines repositories config.
 */
export type RepositoriesConfig = {
  readonly owner: string;
  readonly subscription: string;
};

/**
 * Defines repository config.
 */
export type RepositoryConfig = {
  readonly name: string;
  readonly description: string;
  readonly visibility?: string;
  readonly protected?: boolean;
  readonly topics?: readonly string[];
  readonly homepage?: string;
  readonly enableWiki?: boolean;
  readonly enableDiscussions?: boolean;
  readonly createProject?: boolean;
  readonly rulesets?: RepositoryRulesetsConfig;
  readonly accessPermissions?: RepositoryAccessPermissionsConfig;
};

/**
 * Defines repository rulesets config.
 */
export type RepositoryRulesetsConfig = {
  readonly branch?: RepositoryRulesetConfig;
  readonly tag?: RepositoryRulesetConfig;
};

/**
 * Defines repository branch protections config.
 */
export type RepositoryRulesetConfig = {
  readonly enabled: boolean;
  readonly patterns?: readonly string[];
  readonly restrictCreation?: boolean;
  readonly allowForcePush?: boolean;
  readonly requireConversationResolution?: boolean;
  readonly requireSignedCommits?: boolean;
  readonly requireCodeOwnerReview?: boolean;
  readonly approvingReviewCount?: number;
  readonly requireLastPushApproval?: boolean;
  readonly requireUpdatedBranchBeforeMerge?: boolean;
  readonly requiredChecks?: string[];
  readonly allowBypass?: boolean;
  readonly allowBypassIntegrations?: readonly number[];
};

/**
 * Defines repository access permissions config.
 */
export type RepositoryAccessPermissionsConfig = {
  readonly google?: RepositoryGoogleAccessConfig;
  readonly aws?: RepositoryAwsAccessConfig;
  readonly pulumi?: boolean;
};

/**
 * Defines repository common cloud access config.
 */
export type RepositoryCommonCloudAccessConfig = {
  readonly region?: string;
  readonly iamPermissions?: readonly string[];
};

/**
 * Defines repository Google access config.
 */
export type RepositoryGoogleAccessConfig = RepositoryCommonCloudAccessConfig & {
  readonly project: string;
  readonly linkedProjects?: readonly string[];
  readonly enabledServices?: readonly string[];
  readonly hmacKey?: boolean;
};

/**
 * Defines repository AWS access config.
 */
export type RepositoryAwsAccessConfig = RepositoryCommonCloudAccessConfig & {
  readonly account: string;
};
