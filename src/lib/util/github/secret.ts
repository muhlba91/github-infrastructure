import * as github from '@pulumi/github';
import { Output } from '@pulumi/pulumi';

/**
 * Stores a value in GitHub Actions secrets.
 *
 * @param {string} key the key
 * @param {Output<string>} value the value
 * @param {github.Repository} repository the repository
 */
export const writeToGitHubActionsSecret = (
  key: string,
  value: Output<string>,
  repository: github.Repository,
) => {
  repository.name.apply(
    (repositoryName) =>
      new github.ActionsSecret(
        `github-actions-secret-${repositoryName}-${key}`,
        {
          repository: repositoryName,
          secretName: key,
          plaintextValue: value,
          destroyOnDrift: false,
        },
        {
          dependsOn: [repository],
        },
      ),
  );
};
