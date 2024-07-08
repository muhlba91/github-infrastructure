import * as pulumiservice from '@pulumi/pulumiservice';
import * as vault from '@pulumi/vault';

import { StringMap } from '../../model/map';
import { repositories } from '../configuration';
import { writeToVault } from '../util/vault/secret';
import { vaultProvider } from '../vault';

/**
 * Creates all Pulumi related infrastructure.
 *
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @returns {string[]} the repositories which requested an access token
 */
export const configurePulumi = (
  vaultStores: StringMap<vault.Mount>,
): string[] => {
  const repos = repositories
    .filter((repo) => repo.accessPermissions?.pulumi)
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
  const accessToken = new pulumiservice.AccessToken(
    `pulumi-access-token-${repository}`,
    {
      description: `GitHub Repository: ${repository}`,
    },
    {},
  );

  writeToVault(
    'pulumi',
    accessToken.value.apply((token) =>
      JSON.stringify({
        access_token: token ?? '',
      }),
    ),
    vaultProvider,
    vaultStores[repository],
  );
};
