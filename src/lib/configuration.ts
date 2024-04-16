import {
  Config,
  getOrganization,
  getStack,
  StackReference,
} from '@pulumi/pulumi';

import { AwsConfig } from '../model/config/aws';
import { GcpConfig } from '../model/config/google';
import { RepositoriesConfig } from '../model/config/repository';
import { TailscaleConfig } from '../model/config/tailscale';
import { VaultConfig } from '../model/config/vault';

import { getOrDefault } from './util/get_or_default';
import { parseRepositoriesFromFiles } from './util/repository';

export const environment = getStack();

const config = new Config();
export const repositoriesConfig =
  config.requireObject<RepositoriesConfig>('repositories');
export const awsConfig = config.requireObject<AwsConfig>('aws');
export const gcpConfig = config.requireObject<GcpConfig>('google');
export const vaultConfig = config.requireObject<VaultConfig>('vault');
export const tailscaleConfig =
  config.requireObject<TailscaleConfig>('tailscale');

export const allowRepositoryDeletion =
  getOrDefault(process.env.ALLOW_REPOSITORY_DELETION?.toLowerCase(), 'false') ==
  'true';

export const repositories = parseRepositoriesFromFiles('./assets/repositories');

const vaultStack = new StackReference(
  `${getOrganization()}/muehlbachler-hashicorp-vault-infrastructure/${environment}`,
);
const vaultStackVault = vaultStack.getOutput('vault');
export const vaultConnectionConfig = vaultStackVault.apply((output) => ({
  address: vaultConfig.address,
  token: output?.keys?.rootToken as string,
}));
export const hasVaultConnection = vaultConnectionConfig.token.apply(
  (token) => vaultConfig.enabled && token != undefined,
);

export const commonLabels = {
  environment: environment,
};
