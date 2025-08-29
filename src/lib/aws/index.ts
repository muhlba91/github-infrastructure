import * as aws from '@pulumi/aws';
import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { RepositoryConfig } from '../../model/config/repository';
import { AwsRepositoryAccountData } from '../../model/data/aws';
import { StringMap } from '../../model/map';
import { awsConfig, repositories } from '../configuration';
import { uniqueFilter } from '../util/filter';

import { createAccountIam } from './iam';
import { createAccountGitHubOidc } from './oidc';

const DEFAULT_PERMISSIONS = ['iam:*', 's3:*', 'kms:*'];

/**
 * Creates all AWS related infrastructure.
 *
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @return {StringMap<string[]>} the configured AWS accounts
 */
export const configureAwsAccounts = (
  vaultStores: StringMap<vault.Mount>,
): StringMap<string[]> => {
  const providers = Object.fromEntries(
    Object.entries(awsConfig.account).map(([account, config]) => [
      account,
      new aws.Provider(`aws-provider-${account}`, {
        assumeRoles: [
          {
            roleArn: config.roleArn,
            externalId: config.externalId,
          },
        ],
      }),
    ]),
  );

  const awsRepositoryAccounts = repositories
    .filter((repo) => repo.accessPermissions?.aws?.account)
    .filter(filterRepositoryByAllowedAccounts)
    .map((repo) => ({
      repository: repo.name,
      id: repo.accessPermissions?.aws?.account ?? '',
      region: repo.accessPermissions?.aws?.region ?? awsConfig.defaultRegion,
      iamPermissions: DEFAULT_PERMISSIONS.concat(
        repo.accessPermissions?.aws?.iamPermissions ?? [],
      ),
    }));

  const identityProviderArns = Object.fromEntries(
    awsRepositoryAccounts
      .map((repositoryAccount) => repositoryAccount.id)
      .filter(uniqueFilter)
      .map((repositoryAccount) => [
        repositoryAccount,
        createAccountGitHubOidc(
          repositoryAccount,
          providers[repositoryAccount],
        ),
      ]),
  );
  awsRepositoryAccounts.forEach((repositoryAccount) =>
    configureAccount(
      repositoryAccount,
      providers,
      identityProviderArns[repositoryAccount.id],
      vaultStores,
    ),
  );

  return awsRepositoryAccounts
    .map((repostoryAccount) => ({
      id: repostoryAccount.id,
      repository: repostoryAccount.repository,
    }))
    .reduce<StringMap<string[]>>((accounts, account) => {
      const group = (accounts[account.id] ?? []).concat(account.repository);
      return {
        ...accounts,
        [account.id]: group,
      };
    }, {});
};

/**
 * Configures an AWS account.
 *
 * @param {AwsRepositoryAccountData} account the AWS account
 * @param {StringMap<aws.Provider>} providers the providers for all projects
 * @param {Output<string>} identityProviderArn the identity provider to assign permission for
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 */
const configureAccount = (
  account: AwsRepositoryAccountData,
  providers: StringMap<aws.Provider>,
  identityProviderArn: Output<string>,
  vaultStores: StringMap<vault.Mount>,
) => {
  createAccountIam(account, identityProviderArn, providers, vaultStores);
};

/**
 * Filters the repository by the configured accounts.
 *
 * @param {RepositoryConfig} repository the repository
 * @returns {boolean} true is all accounts are configured; false otherwise
 */
const filterRepositoryByAllowedAccounts = (
  repository: RepositoryConfig,
): boolean => {
  const accounts = Object.keys(awsConfig.account);

  const mainAccount = repository.accessPermissions?.aws?.account;
  if (mainAccount == undefined || !accounts.includes(mainAccount.toString())) {
    console.error(
      `[aws][${repository.name}][${mainAccount}] the repository references an unconfigured account`,
    );
    return false;
  }

  return true;
};
