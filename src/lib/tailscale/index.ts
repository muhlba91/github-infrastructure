import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { StringMap } from '../../model/map';
import { repositories, tailscaleConfig } from '../configuration';
import { writeToVault } from '../util/vault/secret';
import { vaultProvider } from '../vault';

/**
 * Creates all Tailscale related infrastructure.
 *
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @returns {string[]} the repositories which requested an access token
 */
export const configureTailscale = (
  vaultStores: StringMap<vault.Mount>,
): string[] => {
  const repos = repositories
    .filter((repo) => repo.accessPermissions?.tailscale)
    .map((repo) => repo.name);

  repos.forEach((repository) => configureRepository(repository, vaultStores));

  return repos;
};

/**
 * Configures a repository for Pulumi.
 *
 * @param {string} repository the repository
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 */
const configureRepository = (
  repository: string,
  vaultStores: StringMap<vault.Mount>,
) => {
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
