/**
 * Defines an AWS account for a repository.
 */
export type AwsRepositoryAccountData = {
  readonly repository: string;
  readonly id: string;
  readonly region: string;
  readonly iamPermissions: readonly string[];
};
