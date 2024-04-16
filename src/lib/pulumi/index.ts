import * as pulumiservice from '@pulumi/pulumiservice';
import * as vault from '@pulumi/vault';
import * as doppler from '@pulumiverse/doppler';

import { StringMap } from '../../model/map';
import { repositories } from '../configuration';
import { writeToDoppler } from '../util/doppler/secret';
import { writeToVault } from '../util/vault/secret';
import { vaultProvider } from '../vault';

/**
 * Creates all Pulumi related infrastructure.
 *
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @returns {string[]} the repositories which requested an access token
 */
export const configurePulumi = (
  dopplerEnvironments: StringMap<doppler.Environment>,
  vaultStores: StringMap<vault.Mount>,
): string[] => {
  const repos = repositories
    .filter((repo) => repo.accessPermissions?.pulumi)
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
