import * as github from '@pulumi/github';
import { RunError } from '@pulumi/pulumi';

import { RepositoryConfig } from '../../model/config/repository';
import { StringMap } from '../../model/map';
import {
  allowRepositoryDeletion,
  repositories,
  repositoriesConfig,
  stack,
} from '../configuration';
import { getOrDefault } from '../util/get_or_default';
import { isPrivate } from '../util/github/repository';
import { hasSubscription } from '../util/github/subscription';

import { createRepositoryProject } from './project';
import { createRepositoryRulesets } from './ruleset';

const DEFAULT_GITHUB_PAGES_BRANCH = 'main';

/**
 * Creates all GitHub repositories.
 *
 * @returns {StringMap<github.Repository>} the configured repositories
 */
export const createRepositories = (): StringMap<github.Repository> =>
  Object.fromEntries(
    repositories.map((repository) => [
      repository.name,
      createRepository(repository),
    ]),
  );

/**
 * Creates a GitHub repository.
 *
 * @param {RepositoryConfig} config the repository configuration
 * @returns {string} the configured repository
 */
const createRepository = (config: RepositoryConfig): github.Repository => {
  const manageLifecycle = getOrDefault(config.manageLifecycle, true);

  const owner = repositoriesConfig.owner;
  const resourceName = `github-repo-${owner}-${config.name}`;

  if (!manageLifecycle) {
    stack.getOutput('repositories').apply((repos) => {
      if (!repos[config.name]) {
        // eslint-disable-next-line functional/no-throw-statements
        throw new RunError(
          `[ERROR] repository '${config.name}' is not imported yet! Please import it using the following command and re-run Pulumi: pulumi import github:index/repository:Repository ${resourceName} ${owner}/${config.name}`,
        );
      }
    });
  }

  const repo = new github.Repository(
    resourceName,
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
      archiveOnDestroy: manageLifecycle ? config.protected : false,
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
              branch: config.pagesBranch ?? DEFAULT_GITHUB_PAGES_BRANCH,
              path: '/',
            },
          },
      squashMergeCommitMessage: 'COMMIT_MESSAGES',
      squashMergeCommitTitle: 'COMMIT_OR_PR_TITLE',
      vulnerabilityAlerts: true,
      securityAndAnalysis: isPrivate(config)
        ? undefined
        : {
            secretScanning: {
              status: 'enabled',
            },
            secretScanningPushProtection: {
              status: 'enabled',
            },
          },
    },
    {
      protect: manageLifecycle || !allowRepositoryDeletion,
      retainOnDelete: manageLifecycle || !allowRepositoryDeletion,
      ignoreChanges: ['securityAndAnalysis', 'template'],
    },
  );

  if ((hasSubscription() || !isPrivate(config)) && config.rulesets) {
    createRepositoryRulesets(owner, config.name, config.rulesets, repo);
  }

  if (config.createProject) {
    createRepositoryProject(owner, config.name, repo);
  }

  return repo;
};
