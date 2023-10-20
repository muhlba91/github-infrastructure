import * as gcp from '@pulumi/gcp';

/**
 * Defines a Google project for a repository.
 */
export type GoogleRepositoryProjectData = {
  readonly repository: string;
  readonly name: string;
  readonly region: string;
  readonly iamPermissions: readonly string[];
  readonly enabledServices: readonly string[];
  readonly linkedProjects?: readonly string[];
  readonly hmacKey?: boolean;
};

/**
 * Defines a Google workload identity pool.
 */
export type GoogleWorkloadIdentityPoolData = {
  readonly workloadIdentityPool: gcp.iam.WorkloadIdentityPool;
  readonly workloadIdentityProvider: gcp.iam.WorkloadIdentityPoolProvider;
};
