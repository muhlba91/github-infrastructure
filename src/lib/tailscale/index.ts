import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';
import * as doppler from '@pulumiverse/doppler';

import { StringMap } from '../../model/map';
import { repositories, tailscaleConfig } from '../configuration';
import { writeToDoppler } from '../util/doppler/secret';
import { writeToVault } from '../util/vault/secret';
import { vaultProvider } from '../vault';

/**
 * Creates all Tailscale related infrastructure.
 *
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @returns {string[]} the repositories which requested an access token
 */
export const configureTailscale = (
  dopplerEnvironments: StringMap<doppler.Environment>,
  vaultStores: StringMap<vault.Mount>,
): string[] => {
  const repos = repositories
    .filter((repo) => repo.accessPermissions?.tailscale)
    .map((repo) => repo.name);

  repos.forEach((repository) =>
    configureRepository(repository, dopplerEnvironments, vaultStores),
  );

  return repos;
};

/**
 * Configures a repository for Pulumi.
 *
 * @param {string} repository the repository
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 */
const configureRepository = (
  repository: string,
  dopplerEnvironments: StringMap<doppler.Environment>,
  vaultStores: StringMap<vault.Mount>,
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

  writeToVault(
    'tailscale',
    Output.create(
      JSON.stringify({
        oauth_client_id: tailscaleConfig.oauth.id,
        oauth_secret: tailscaleConfig.oauth.secret,
      }),
    ),
    vaultProvider,
    vaultStores[repository],
  );
};
