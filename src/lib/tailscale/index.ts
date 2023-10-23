import { Output } from '@pulumi/pulumi';
import * as doppler from '@pulumiverse/doppler';

import { StringMap } from '../../model/map';
import { repositories, tailscaleConfig } from '../configuration';
import { writeToDoppler } from '../util/doppler/secret';

/**
 * Creates all Tailscale related infrastructure.
 *
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @returns {string[]} the repositories which requested an access token
 */
export const configureTailscale = (
  dopplerEnvironments: StringMap<doppler.Environment>,
): string[] => {
  const repos = repositories
    .filter((repo) => repo.accessPermissions?.tailscale)
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
  writeToDoppler(
    'TS_OAUTH_CLIENT_ID',
    Output.create(tailscaleConfig.oauth.id),
    dopplerEnvironments[repository],
  );
  writeToDoppler(
    'TS_OAUTH_SECRET',
    Output.create(tailscaleConfig.oauth.secret),
    dopplerEnvironments[repository],
  );
};
