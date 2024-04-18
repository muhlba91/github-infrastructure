import * as github from '@pulumi/github';
import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { RepositoryConfig } from '../../model/config/repository';
import { repositoriesConfig, vaultConfig } from '../configuration';
import { writeToGitHubActionsSecret } from '../util/github/secret';
import { renderTemplate } from '../util/template';
import { writeToVault } from '../util/vault/secret';

/**
 * Creates a Vault authentication.
 *
 * @param {RepositoryConfig} repository the repository
 * @param {vault.Mount} store the Vault store
 * @param {github.Repository} githubRepository the GitHub repository
 * @param {vault.Provider} provider the Vault provider
 * @returns {vault.jwt.AuthBackendRole} the Vault authentication
 */
export const createAuth = (
  repository: RepositoryConfig,
  store: vault.Mount,
  githubRepository: github.Repository,
  provider: vault.Provider,
): vault.jwt.AuthBackendRole => {
  new vault.Policy(
    `vault-policy-github-${repository.name}`,
    {
      name: `github-${repository.name}`,
      policy: renderTemplate('assets/vault/policy.hcl.tpl', {
        repository: repository.name,
        additionalPaths:
          repository.accessPermissions?.vault?.additionalMounts?.map(
            (mount) => ({
              path: mount.path,
              permissions: mount.permissions
                .map((permission) => `"${permission}"`)
                .join(', '),
            }),
          ) ?? [],
      }),
    },
    {
      provider: provider,
    },
  );

  const vaultAddress =
    repository.accessPermissions?.vault?.address ?? vaultConfig.address;
  const jwtRole = new vault.jwt.AuthBackendRole(
    `vault-jwt-github-role-${repository.name}`,
    {
      backend: 'github',
      roleType: 'jwt',
      roleName: `github-${repository.name}`,
      tokenPolicies: [`github-${repository.name}`],
      tokenTtl: 1 * 60 * 60,
      boundAudiences: [`https://github.com/${repositoriesConfig.owner}`],
      userClaim: 'repository',
      boundClaims: {
        repository: `${repositoriesConfig.owner}/${repository.name}`,
      },
    },
    {
      provider: provider,
    },
  );

  writeToVault(
    'vault',
    jwtRole.roleName.apply((roleName) =>
      JSON.stringify({
        address: vaultAddress,
        role: roleName,
        path: 'github',
      }),
    ),
    provider,
    store,
  );

  writeToGitHubActionsSecret(
    'VAULT_ADDR',
    Output.create(vaultAddress),
    githubRepository,
  );
  writeToGitHubActionsSecret('VAULT_ROLE', jwtRole.roleName, githubRepository);
  writeToGitHubActionsSecret(
    'VAULT_PATH',
    Output.create('github'),
    githubRepository,
  );

  return jwtRole;
};
