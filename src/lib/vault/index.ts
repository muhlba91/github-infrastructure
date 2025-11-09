import * as github from '@pulumi/github';
import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { StringMap } from '../../model/map';
import {
  hasVaultConnection,
  repositories,
  vaultConnectionConfig,
} from '../configuration';
import { uniqueFilter } from '../util/filter';
import { getOrDefault } from '../util/get_or_default';

import { createRepositoryStore, createStore } from './store';

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
): Output<StringMap<vault.Mount>> =>
  hasVaultConnection.apply((hasConnection) => {
    const repos = (hasConnection ? repositories : []).filter(
      (repository) =>
        getOrDefault(repository.manageLifecycle, true) &&
        getOrDefault(repository.accessPermissions?.vault?.enabled, true),
    );

    repos
      .flatMap(
        (repository) =>
          repository.accessPermissions?.vault?.additionalMounts ?? [],
      )
      .filter((mount) => getOrDefault(mount.create, false))
      .map((mount) => mount.path)
      .filter(uniqueFilter)
      .forEach((path) =>
        createStore(path, path, `Secrets for: ${path}`, vaultProvider),
      );

    return Object.fromEntries(
      repos.map((repository) => [
        repository.name,
        createRepositoryStore(
          repository,
          githubRepositories[repository.name],
          vaultProvider,
        ),
      ]),
    );
  });
