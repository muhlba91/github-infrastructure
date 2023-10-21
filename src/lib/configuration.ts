import { Config, getStack } from '@pulumi/pulumi';

import { AwsConfig } from '../model/config/aws';
import { GcpConfig } from '../model/config/google';
import { RepositoriesConfig } from '../model/config/repository';

import { getOrDefault } from './util/get_or_default';
import { parseRepositoriesFromFiles } from './util/repository';

export const environment = getStack();

const config = new Config();
export const repositoriesConfig =
  config.requireObject<RepositoriesConfig>('repositories');
export const awsConfig = config.requireObject<AwsConfig>('aws');
export const gcpConfig = config.requireObject<GcpConfig>('google');

export const allowRepositoryDeletion =
  getOrDefault(process.env.ALLOW_REPOSITORY_DELETION?.toLowerCase(), 'false') ==
  'true';

export const repositories = parseRepositoriesFromFiles('./assets/repositories');

export const commonLabels = {
  environment: environment,
};
