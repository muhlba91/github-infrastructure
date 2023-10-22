import * as github from '@pulumi/github';

/**
 * Creates GitHub repository projects.
 *
 * @param {string} owner the repository owner
 * @param {string} name the repository name
 * @param {github.Repository} repository the GitHub repository
 */
export const createRepositoryProject = (
  owner: string,
  name: string,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  repository: github.Repository,
) => {
  console.warn(
    `[github][project][${name}] a project was requested but creation is currently not supported by GitHub`,
  );
};
