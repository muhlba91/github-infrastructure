import { StringMap } from '../map';

/**
 * Defines AWS config.
 */
export interface AwsConfig {
  readonly defaultRegion: string;
  readonly account: StringMap<AwsAccountConfig>;
}

/**
 * Defines AWS account config.
 */
export interface AwsAccountConfig {
  readonly externalId: string;
  readonly roleArn: string;
}
