import * as aws from '@pulumi/aws';
import { Output } from '@pulumi/pulumi';

import { commonLabels } from '../configuration';

/**
 * Creates GitHub OIDC for an AWS account.
 *
 * @param {string} account the AWS account
 * @param {aws.Provider} provider the AWS provider
 * @returns {Output<string>} the identity provider ARN
 */
export const createAccountGitHubOidc = (
  account: string,
  provider: aws.Provider,
): Output<string> => {
  const identityProvider = new aws.iam.OpenIdConnectProvider(
    `aws-iam-identity-provider-${account}`,
    {
      url: 'https://token.actions.githubusercontent.com',
      clientIdLists: ['sts.amazonaws.com'],
      thumbprintLists: ['ffffffffffffffffffffffffffffffffffffffff'],
      tags: {
        ...commonLabels,
        purpose: 'github-actions',
      },
    },
    {
      provider: provider,
    },
  );

  return identityProvider.arn;
};
