import * as gcp from '@pulumi/gcp';
import { interpolate, Resource } from '@pulumi/pulumi';

import { GoogleWorkloadIdentityPoolData } from '../../model/data/google';
import { createRandomString } from '../util/random';

/**
 * Creates GitHub OIDC for a Google project.
 *
 * @param {string} project the Google project
 * @param {gcp.Provider} provider the Google provider
 * @param {Resource[]} dependencies the Pulumi dependencies
 * @returns {GoogleWorkloadIdentityPoolData} the workload identity pool
 */
export const createProjectGitHubOidc = (
  project: string,
  provider: gcp.Provider,
  dependencies: Resource[],
): GoogleWorkloadIdentityPoolData => {
  const poolPostfix = createRandomString(
    `gcp-iam-identity-pool-${project}`,
    {},
  ).result.apply((id) => id.toLowerCase());

  const workloadIdentityPool = new gcp.iam.WorkloadIdentityPool(
    `gcp-iam-identity-pool-${project}`,
    {
      workloadIdentityPoolId: interpolate`github-${poolPostfix}`,
      displayName: 'GitHub Identity Pool',
      description: 'Workload Identity pool to federate GitHub repositories',
      project: project,
    },
    {
      provider: provider,
      dependsOn: dependencies,
    },
  );
  const workloadIdentityProvider = new gcp.iam.WorkloadIdentityPoolProvider(
    `gcp-iam-identity-provider-${project}`,
    {
      workloadIdentityPoolId: workloadIdentityPool.workloadIdentityPoolId,
      workloadIdentityPoolProviderId: interpolate`github-actions-${poolPostfix}`,
      displayName: 'GitHub Actions Provider',
      description: 'Workload Identity Provider to federate GitHub Actions',
      oidc: {
        issuerUri: 'https://token.actions.githubusercontent.com',
      },
      attributeMapping: {
        'google.subject': 'assertion.sub',
        'attribute.actor': 'assertion.actor',
        'attribute.repository_owner': 'assertion.repository_owner',
        'attribute.repository': 'assertion.repository',
      },
      project: project,
    },
    {
      provider: provider,
      dependsOn: dependencies,
    },
  );

  return {
    workloadIdentityPool: workloadIdentityPool,
    workloadIdentityProvider: workloadIdentityProvider,
  };
};
