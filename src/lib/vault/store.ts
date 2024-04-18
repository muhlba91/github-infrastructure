import * as github from '@pulumi/github';
import * as vault from '@pulumi/vault';

import { RepositoryConfig } from '../../model/config/repository';
import { repositoriesConfig } from '../configuration';

import { createAuth } from './auth';

/**
 * Creates a Vault store for a repository.
 *
 * @param {RepositoryConfig} repository the repository
 * @param {github.Repository} githubRepository the GitHub repository
 * @param {vault.Provider} provider the Vault provider
 * @returns {vault.Mount} the Vault store
 */
export const createRepositoryStore = (
  repository: RepositoryConfig,
  githubRepository: github.Repository,
  provider: vault.Provider,
): vault.Mount => {
  const mount = createStore(
    repository.name,
    `github-${repository.name}`,
    `GitHub repository: ${repositoriesConfig.owner}/${repository.name}`,
    provider,
  );

  createAuth(repository, mount, githubRepository, provider);

  return mount;
};

/**
 * Creates a Vault store.
 *
 * @param {string} name the name
 * @param {string} path the path
 * @param {string} description the description
 * @param {vault.Provider} provider the Vault provider
 * @returns {vault.Mount} the Vault store
 */
export const createStore = (
  name: string,
  path: string,
  description: string,
  provider: vault.Provider,
): vault.Mount =>
  new vault.Mount(
    `vault-store-${name}`,
    {
      path: path,
      type: 'kv',
      options: {
        version: '2',
      },
      description: description,
    },
    {
      provider: provider,
    },
  );
