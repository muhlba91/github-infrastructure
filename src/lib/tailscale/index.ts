import { all } from '@pulumi/pulumi';
import * as tailscale from '@pulumi/tailscale';
import * as vault from '@pulumi/vault';

import { StringMap } from '../../model/map';
import { repositories } from '../configuration';
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
  const oauthClient = new tailscale.OauthClient(
    `tailscale-oauth-client-${repository}`,
    {
      description: repository.substring(0, 50),
      scopes: ['all'],
    },
  );

  writeToVault(
    'tailscale',
    all([oauthClient.id, oauthClient.key]).apply(([id, key]) =>
      JSON.stringify({
        oauth_client_id: id,
        oauth_secret: key,
      }),
    ),
    vaultProvider,
    vaultStores[repository],
  );
};
