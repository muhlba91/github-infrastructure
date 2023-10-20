import * as pulumiservice from '@pulumi/pulumiservice';

import { repositories } from '../configuration';
import { writeToDoppler } from '../util/doppler/secret';

/**
 * Creates all Pulumi related infrastructure.
 *
 * @returns {string[]} the repositories which requested an access token
 */
export const configurePulumi = (): string[] => {
  const repos = repositories
    .filter((repo) => repo.accessPermissions?.pulumi)
    .map((repo) => repo.name);

  repos.forEach(configureRepository);

  return repos;
};

/**
 * Configures a repository for Pulumi.
 *
 * @param {string} repository the repository
 */
const configureRepository = (repository: string) => {
  const accessToken = new pulumiservice.AccessToken(
    `pulumi-access-token-${repository}`,
    {
      description: `GitHub Repository: ${repository}`,
    },
    {},
  );

  writeToDoppler(
    'PULUMI_ACCESS_TOKEN',
    accessToken.value.apply((token) => token ?? ''),
    repository,
  );
};
