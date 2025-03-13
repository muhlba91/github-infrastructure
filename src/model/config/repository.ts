import { StringMap } from '../map';

/**
 * Defines repositories config.
 */
export interface RepositoriesConfig {
  readonly owner: string;
  readonly subscription: string;
}

/**
 * Defines repository config.
 */
export interface RepositoryConfig {
  readonly name: string;
  readonly description: string;
  readonly visibility?: string;
  readonly protected?: boolean;
  readonly topics?: readonly string[];
  readonly homepage?: string;
  readonly enableWiki?: boolean;
  readonly enableDiscussions?: boolean;
  readonly createProject?: boolean;
  readonly pagesBranch?: string;
  readonly rulesets?: RepositoryRulesetsConfig;
  readonly accessPermissions?: RepositoryAccessPermissionsConfig;
}

/**
 * Defines repository rulesets config.
 */
export interface RepositoryRulesetsConfig {
  readonly branch?: RepositoryRulesetConfig;
  readonly tag?: RepositoryRulesetConfig;
}

/**
 * Defines repository branch protections config.
 */
export interface RepositoryRulesetConfig {
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
  readonly enableMergeQueue?: boolean;
  readonly requiredChecks?: string[];
  readonly allowBypass?: boolean;
  readonly allowBypassIntegrations?: readonly number[];
  readonly enableGitstreamIntegration?: boolean;
  readonly enableWipIntegration?: boolean;
}

/**
 * Defines repository access permissions config.
 */
export interface RepositoryAccessPermissionsConfig {
  readonly tailscale?: boolean;
  readonly vault?: RepositoryVaultAccessPermissionsConfig;
  readonly google?: RepositoryGoogleAccessConfig;
  readonly aws?: RepositoryAwsAccessConfig;
}

/**
 * Defines repository vault access permissions config.
 */
export interface RepositoryVaultAccessPermissionsConfig {
  readonly enabled: boolean;
  readonly address?: string;
  readonly additionalMounts?: readonly RepositoryVaultAdditionalMountAccessPermissionsConfig[];
}

/**
 * Defines repository vault additional mount access permissions config.
 */
export interface RepositoryVaultAdditionalMountAccessPermissionsConfig {
  readonly path: string;
  readonly create?: boolean;
  readonly permissions: readonly string[];
}

/**
 * Defines repository common cloud access config.
 */
export interface RepositoryCommonCloudAccessConfig {
  readonly region?: string;
  readonly iamPermissions?: readonly string[];
}

/**
 * Defines repository Google access config.
 */
export type RepositoryGoogleAccessConfig = RepositoryCommonCloudAccessConfig & {
  readonly project: string;
  readonly linkedProjects?: StringMap<RepositoryGoogleLinkedAccessConfig>;
  readonly enabledServices?: readonly string[];
  readonly hmacKey?: boolean;
};

/**
 * Defines repository linked Google access config.
 */
export interface RepositoryGoogleLinkedAccessConfig {
  readonly accessLevel: string;
  readonly iamPermissions?: readonly string[];
}

/**
 * Defines repository AWS access config.
 */
export type RepositoryAwsAccessConfig = RepositoryCommonCloudAccessConfig & {
  readonly account: string;
};
