import { StringMap } from '../map';

/**
 * Defines AWS config.
 */
export type AwsConfig = {
  readonly defaultRegion: string;
  readonly account: StringMap<AwsAccountConfig>;
};

/**
 * Defines AWS account config.
 */
export type AwsAccountConfig = {
  readonly externalId: string;
  readonly roleArn: string;
};
