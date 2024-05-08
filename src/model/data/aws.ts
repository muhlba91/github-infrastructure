/**
 * Defines an AWS account for a repository.
 */
export interface AwsRepositoryAccountData {
  readonly repository: string;
  readonly id: string;
  readonly region: string;
  readonly iamPermissions: readonly string[];
}
