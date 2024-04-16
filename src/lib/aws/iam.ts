import * as aws from '@pulumi/aws';
import { interpolate, Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';
import * as doppler from '@pulumiverse/doppler';

import { AwsRepositoryAccountData } from '../../model/data/aws';
import { StringMap } from '../../model/map';
import { commonLabels, repositoriesConfig } from '../configuration';
import { writeToDoppler } from '../util/doppler/secret';
import { createRandomString } from '../util/random';
import { writeToVault } from '../util/vault/secret';
import { vaultProvider } from '../vault';

/**
 * Creates IAM for an AWS account.
 *
 * @param {AwsRepositoryAccountData} account the AWS account
 * @param {Output<string | undefined>} identityProviderArn the identity provider ARN if created
 * @param {StringMap<aws.Provider>} providers the providers for all projects
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 */
export const createAccountIam = (
  account: AwsRepositoryAccountData,
  identityProviderArn: Output<string | undefined>,
  providers: StringMap<aws.Provider>,
  dopplerEnvironments: StringMap<doppler.Environment>,
  vaultStores: StringMap<vault.Mount>,
) => {
  const labels = {
    ...commonLabels,
    purpose: 'github-repository',
    repository: account.repository,
  };

  const ciPostfix = createRandomString(
    `aws-iam-role-ci-${account.repository}-${account.id}`,
    {},
  );
  const truncatedRepository = account.repository.substring(0, 18);

  const ciRole = identityProviderArn.apply(
    (providerArn) =>
      new aws.iam.Role(
        `aws-iam-role-ci-${account.repository}-${account.id}`,
        {
          name: interpolate`ci-${truncatedRepository}-${ciPostfix.result}`,
          description: `GitHub Repository: ${account.repository}`,
          assumeRolePolicy: JSON.stringify({
            Version: '2012-10-17',
            Statement: [
              {
                Action: 'sts:AssumeRoleWithWebIdentity',
                Effect: 'Allow',
                Principal: {
                  Federated: providerArn,
                },
                Condition: {
                  StringEquals: {
                    'token.actions.githubusercontent.com:aud':
                      'sts.amazonaws.com',
                  },
                  StringLike: {
                    'token.actions.githubusercontent.com:sub': `repo:${repositoriesConfig.owner}/${account.repository}:*`,
                  },
                },
              },
            ],
          }),
          tags: labels,
        },
        {
          provider: providers[account.id],
        },
      ),
  );

  const policy = new aws.iam.Policy(
    `aws-iam-role-ci-policy-${account.repository}-${account.id}`,
    {
      name: interpolate`ci-${truncatedRepository}-${ciPostfix.result}`,
      description: `GitHub Repository: ${account.repository}`,
      policy: JSON.stringify({
        Version: '2012-10-17',
        Statement: [
          {
            Action: account.iamPermissions,
            Effect: 'Allow',
            Resource: '*',
          },
        ],
      }),
      tags: labels,
    },
    {
      provider: providers[account.id],
    },
  );

  new aws.iam.RolePolicyAttachment(
    `aws-iam-role-ci-policy-attachment-${account.repository}-${account.id}`,
    {
      role: ciRole.name,
      policyArn: policy.arn,
    },
    {
      provider: providers[account.id],
      dependsOn: [ciRole, policy],
    },
  );

  writeToDoppler(
    'AWS_IDENTITY_ROLE_ARN',
    ciRole.arn,
    dopplerEnvironments[account.repository],
  );
  writeToDoppler(
    'AWS_REGION',
    Output.create(account.region),
    dopplerEnvironments[account.repository],
  );

  writeToVault(
    'aws',
    ciRole.arn.apply((ciRoleArn) =>
      JSON.stringify({
        identity_role_arn: ciRoleArn,
        region: account.region,
      }),
    ),
    vaultProvider,
    vaultStores[account.repository],
  );
};
