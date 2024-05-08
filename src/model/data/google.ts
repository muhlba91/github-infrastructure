import * as gcp from '@pulumi/gcp';

import { StringMap } from '../map';

/**
 * Defines a Google project for a repository.
 */
export interface GoogleRepositoryProjectData {
  readonly repository: string;
  readonly name: string;
  readonly region: string;
  readonly iamPermissions: readonly string[];
  readonly enabledServices: readonly string[];
  readonly linkedProjects?: StringMap<GoogleRepositoryLinkedProjectData>;
  readonly hmacKey?: boolean;
}

/**
 * Defines a linked Google project.
 */
export interface GoogleRepositoryLinkedProjectData {
  readonly accessLevel: string;
  readonly iamPermissions?: readonly string[];
}

/**
 * Defines a Google workload identity pool.
 */
export interface GoogleWorkloadIdentityPoolData {
  readonly workloadIdentityPool: gcp.iam.WorkloadIdentityPool;
  readonly workloadIdentityProvider: gcp.iam.WorkloadIdentityPoolProvider;
}
