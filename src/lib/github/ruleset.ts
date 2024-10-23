import * as github from '@pulumi/github';

import {
  RepositoryRulesetConfig,
  RepositoryRulesetsConfig,
} from '../../model/config/repository';
import { getOrDefault } from '../util/get_or_default';

const DEFAULT_BRANCH_RULESET_PATTERNS = ['~DEFAULT_BRANCH'];
const DEFAULT_TAG_RULESET_PATTERNS = [''];

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
        requiredStatusChecks: config.requiredChecks
          ? {
              requiredChecks:
                config.requiredChecks?.map((check) => ({
                  context: check,
                  integrationId: 15368, // GitHub Actions
                })) ?? [],
              strictRequiredStatusChecksPolicy: getOrDefault(
                config.requireUpdatedBranchBeforeMerge,
                true,
              ),
            }
          : undefined,
        update: false,
        updateAllowsFetchAndMerge: false,
      },
      bypassActors: getOrDefault(config.allowBypass, true)
        ? [
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
          ].concat(
            config.allowBypassIntegrations?.map((integration) => ({
              actorId: integration,
              actorType: 'Integration',
              bypassMode: 'always',
            })) ?? [],
          )
        : [],
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
