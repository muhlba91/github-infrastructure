import * as pulumiservice from '@pulumi/pulumiservice';
import * as doppler from '@pulumiverse/doppler';

import { StringMap } from '../../model/map';
import { repositories } from '../configuration';
import { writeToDoppler } from '../util/doppler/secret';

/**
 * Creates all Pulumi related infrastructure.
 *
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @returns {string[]} the repositories which requested an access token
 */
export const configurePulumi = (
  dopplerEnvironments: StringMap<doppler.Environment>,
): string[] => {
  const repos = repositories
    .filter((repo) => repo.accessPermissions?.pulumi)
    .map((repo) => repo.name);

  repos.forEach((repository) =>
    configureRepository(repository, dopplerEnvironments),
  );

  return repos;
};

/**
 * Configures a repository for Pulumi.
 *
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @param {string} repository the repository
 */
const configureRepository = (
  repository: string,
  dopplerEnvironments: StringMap<doppler.Environment>,
) => {
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
    dopplerEnvironments[repository],
  );
};
