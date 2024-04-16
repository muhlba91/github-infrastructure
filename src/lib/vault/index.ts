import * as github from '@pulumi/github';
import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { StringMap } from '../../model/map';
import {
  hasVaultConnection,
  repositories,
  vaultConnectionConfig,
} from '../configuration';
import { getOrDefault } from '../util/get_or_default';

import { createStore } from './store';

export const vaultProvider = new vault.Provider('vault', {
  address: vaultConnectionConfig.address,
  token: vaultConnectionConfig.token,
});

/**
 * Creates all Vault secret stores.
 *
 * @returns {Output<StringMap<vault.Mount>>} all Vault secret stores
 */
export const configureVaultStores = (
  githubRepositories: StringMap<github.Repository>,
): Output<StringMap<vault.Mount>> => {
  return hasVaultConnection.apply((hasConnection) =>
    Object.fromEntries(
      (hasConnection ? repositories : [])
        .filter(
          (repository) =>
            getOrDefault(repository.accessPermissions?.vault?.enabled, true),
        )
        .map((repository) => [
          repository.name,
          createStore(
            repository,
            githubRepositories[repository.name],
            vaultProvider,
          ),
        ]),
    ),
  );
};
