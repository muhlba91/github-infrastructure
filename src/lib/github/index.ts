import * as github from '@pulumi/github';

import { RepositoryConfig } from '../../model/config/repository';
import {
  allowRepositoryDeletion,
  repositories,
  repositoriesConfig,
} from '../configuration';
import { getOrDefault } from '../util/get_or_default';
import { isPrivate } from '../util/github/repository';
import { hasSubscription } from '../util/github/subscription';

import { createRepositoryProject } from './project';
import { createRepositoryRulesets } from './ruleset';

/**
 * Creates all GitHub repositories.
 *
 * @returns {string[]} the configured repositories
 */
export const createRepositories = (): string[] =>
  repositories.map(createRepository);

/**
 * Creates a GitHub repository.
 *
 * @param {RepositoryConfig} config the repository configuration
 * @returns {string} the configured repository
 */
const createRepository = (config: RepositoryConfig): string => {
  const owner = repositoriesConfig.owner;
  const repo = new github.Repository(
    `github-repo-${owner}-${config.name}`,
    {
      name: config.name,
      description: config.description,
      hasDiscussions: config.enableDiscussions,
      hasWiki: config.enableWiki,
      homepageUrl: config.homepage,
      topics: config.topics?.map((topic) => topic).sort(),
      visibility: getOrDefault(config.visibility, 'public'),
      allowAutoMerge: false,
      allowMergeCommit: false,
      allowRebaseMerge: true,
      allowSquashMerge: false,
      allowUpdateBranch: true,
      archived: false,
      archiveOnDestroy: config.protected,
      autoInit: false,
      deleteBranchOnMerge: true,
      hasDownloads: true,
      hasIssues: true,
      hasProjects: true,
      mergeCommitMessage: 'PR_TITLE',
      mergeCommitTitle: 'MERGE_MESSAGE',
      pages: isPrivate(config)
        ? undefined
        : {
            buildType: 'workflow',
            source: {
              branch: config.pagesBranch ?? 'main',
              path: '/',
            },
          },
      squashMergeCommitMessage: 'COMMIT_MESSAGES',
      squashMergeCommitTitle: 'COMMIT_OR_PR_TITLE',
      vulnerabilityAlerts: true,
      securityAndAnalysis: {
        secretScanning: {
          status: 'enabled',
        },
        secretScanningPushProtection: {
          status: 'enabled',
        },
      },
    },
    {
      protect: !allowRepositoryDeletion,
      retainOnDelete: !allowRepositoryDeletion,
      ignoreChanges: ['securityAndAnalysis'],
    },
  );

  if ((hasSubscription() || !isPrivate(config)) && config.rulesets) {
    createRepositoryRulesets(owner, config.name, config.rulesets, repo);
  }

  if (config.createProject) {
    createRepositoryProject(owner, config.name, repo);
  }

  return config.name;
};
