import * as github from '@pulumi/github';

import {
  RepositoryRulesetConfig,
  RepositoryRulesetsConfig,
} from '../../model/config/repository';
import { getOrDefault } from '../util/get_or_default';

const DEFAULT_BRANCH_RULESET_PATTERNS = ['~DEFAULT_BRANCH'];
const DEFAULT_TAG_RULESET_PATTERNS = [''];

const GITHUB_ACTIONS_INTEGRATION_ID = 15368;
const GITSTREAM_INTEGRATION_ID = 230441;
const WIP_INTEGRATION_ID = 3414;

/**
 * Creates GitHub repository rulesets.
 *
 * @param {string} owner the repository owner
 * @param {string} name the repository name
 * @param {RepositoryRulesetsConfig} config the repository rulesets configuration
 * @param {github.Repository} repository the GitHub repository
 */
export const createRepositoryRulesets = (
  owner: string,
  name: string,
  config: RepositoryRulesetsConfig,
  repository: github.Repository,
) => {
  if (config.branch?.enabled) {
    createRepositoryRuleset(
      owner,
      name,
      'branch',
      config.branch,
      DEFAULT_BRANCH_RULESET_PATTERNS,
      repository,
    );
  }
  if (config.tag?.enabled) {
    createRepositoryRuleset(
      owner,
      name,
      'tag',
      config.tag,
      DEFAULT_TAG_RULESET_PATTERNS,
      repository,
    );
  }
};

/**
 * Creates a GitHub repository ruleset.
 *
 * @param {string} owner the repository owner
 * @param {string} name the repository name
 * @param {string} target the ruleset target
 * @param {RepositoryRulesetConfig} config the repository ruleset configuration
 * @param {string[]} defaultPatterns default patterns to use
 * @param {github.Repository} repository the GitHub repository
 */
const createRepositoryRuleset = (
  owner: string,
  name: string,
  target: string,
  config: RepositoryRulesetConfig,
  defaultPatterns: readonly string[],
  repository: github.Repository,
) => {
  new github.RepositoryRuleset(
    `github-repository-ruleset-${target}-${owner}-${name}`,
    {
      repository: name,
      target: 'branch',
      enforcement: 'active',
      rules: {
        creation: getOrDefault(config.restrictCreation, true),
        deletion: true,
        nonFastForward: !getOrDefault(config.allowForcePush, false),
        pullRequest: {
          dismissStaleReviewsOnPush: true,
          requireCodeOwnerReview: getOrDefault(
            config.requireCodeOwnerReview,
            false,
          ),
          requiredApprovingReviewCount: getOrDefault(
            config.approvingReviewCount,
            0,
          ),
          requiredReviewThreadResolution: getOrDefault(
            config.requireConversationResolution,
            true,
          ),
          requireLastPushApproval: getOrDefault(
            config.requireLastPushApproval,
            true,
          ),
        },
        requiredDeployments: {
          requiredDeploymentEnvironments: [],
        },
        requiredLinearHistory: true,
        requiredSignatures: getOrDefault(config.requireSignedCommits, false),
        requiredStatusChecks: computeRequiredChecks(config),
        update: false,
        updateAllowsFetchAndMerge: false,
      },
      bypassActors: computeBypassActors(config),
      conditions: {
        refName: {
          excludes: [],
          includes: defaultPatterns.concat(config.patterns ?? []),
        },
      },
    },
    {
      dependsOn: [repository],
    },
  );
};

/**
 * Computes the required checks for a repository ruleset.
 *
 * @param {RepositoryRulesetConfig} config the repository ruleset configuration
 * @returns {github.types.input.RepositoryRulesetRulesRequiredStatusChecks | undefined} the required checks configuration
 */
const computeRequiredChecks = (
  config: RepositoryRulesetConfig,
):
  | github.types.input.RepositoryRulesetRulesRequiredStatusChecks
  | undefined => {
  if (
    !config.requiredChecks &&
    !getOrDefault(config.enableGitstreamIntegration, true)
  ) {
    return undefined;
  }

  const requiredChecks =
    config.requiredChecks?.map((check) => ({
      context: check,
      integrationId: GITHUB_ACTIONS_INTEGRATION_ID,
    })) ?? [];

  const gitstreamIntegration = getOrDefault(
    config.enableGitstreamIntegration,
    true,
  )
    ? [
        {
          context: 'gitStream.cm',
          integrationId: GITSTREAM_INTEGRATION_ID,
        },
      ]
    : [];

  const wipIntegration = getOrDefault(config.enableWipIntegration, true)
    ? [
        {
          context: 'WIP',
          integrationId: WIP_INTEGRATION_ID,
        },
      ]
    : [];

  return {
    requiredChecks: requiredChecks
      .concat(gitstreamIntegration)
      .concat(wipIntegration)
      .sort((a, b) => a.context.localeCompare(b.context)),
    strictRequiredStatusChecksPolicy: getOrDefault(
      config.requireUpdatedBranchBeforeMerge,
      true,
    ),
  };
};

/**
 * Computes the bypass actors for a repository ruleset.
 *
 * @param {RepositoryRulesetConfig} config the repository ruleset configuration
 * @returns {github.types.input.RepositoryRulesetBypassActor[]} the bypass actors
 */
const computeBypassActors = (
  config: RepositoryRulesetConfig,
): github.types.input.RepositoryRulesetBypassActor[] => {
  if (!getOrDefault(config.allowBypass, true)) {
    return [];
  }

  const bypassActors = [
    {
      actorId: 2, // maintainer
      actorType: 'RepositoryRole',
      bypassMode: 'pull_request',
    },
    {
      actorId: 5, // admin
      actorType: 'RepositoryRole',
      bypassMode: 'always',
    },
  ];

  const bypassIntegrations =
    config.allowBypassIntegrations?.map((integration) => ({
      actorId: integration,
      actorType: 'Integration',
      bypassMode: 'always',
    })) ?? [];

  const gitstreamIntegration = getOrDefault(
    config.enableGitstreamIntegration,
    true,
  )
    ? [
        {
          actorId: GITSTREAM_INTEGRATION_ID,
          actorType: 'Integration',
          bypassMode: 'always',
        },
      ]
    : [];

  return (
    getOrDefault(config.allowBypass, true)
      ? bypassActors.concat(bypassIntegrations)
      : []
  )
    .concat(gitstreamIntegration)
    .sort((a, b) => a.actorId - b.actorId);
};
