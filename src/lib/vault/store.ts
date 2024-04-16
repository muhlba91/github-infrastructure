import * as github from '@pulumi/github';
import * as vault from '@pulumi/vault';

import { RepositoryConfig } from '../../model/config/repository';
import { repositoriesConfig } from '../configuration';

import { createAuth } from './auth';

/**
 * Creates a Vault store.
 *
 * @param {RepositoryConfig} repository the repository
 * @param {github.Repository} githubRepository the GitHub repository
 * @param {vault.Provider} provider the Vault provider
 * @returns {vault.Mount} the Vault store
 */
export const createStore = (
  repository: RepositoryConfig,
  githubRepository: github.Repository,
  provider: vault.Provider,
): vault.Mount => {
  const mount = new vault.Mount(
    `vault-store-${repository.name}`,
    {
      path: `github-${repository.name}`,
      type: 'kv',
      options: {
        version: '2',
      },
      description: `GitHub repository: ${repositoriesConfig.owner}/${repository.name}`,
    },
    {
      provider: provider,
    },
  );

  createAuth(repository, mount, githubRepository, provider);

  return mount;
};
