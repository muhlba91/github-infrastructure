import { configureAwsAccounts } from './lib/aws';
import { awsConfig, gcpConfig, repositories } from './lib/configuration';
import { configureDoppler } from './lib/doppler';
import { createRepositories } from './lib/github';
import { configureGoogleProjects } from './lib/google';
import { configurePulumi } from './lib/pulumi';
import { configureTailscale } from './lib/tailscale';
import { getOrDefault } from './lib/util/get_or_default';
import { configureVaultStores } from './lib/vault';

export = async () => {
  const githubRepositories = createRepositories();

  const dopplerEnvironments = configureDoppler();
  const vaultStores = configureVaultStores(githubRepositories);
  const pulumis = vaultStores.apply((stores) =>
    configurePulumi(dopplerEnvironments, stores),
  );
  const tailscales = vaultStores.apply((stores) =>
    configureTailscale(dopplerEnvironments, stores),
  );
  const projects = vaultStores.apply((stores) =>
    configureGoogleProjects(dopplerEnvironments, stores),
  );
  const accounts = vaultStores.apply((stores) =>
    configureAwsAccounts(dopplerEnvironments, stores),
  );

  return {
    google: {
      allowed: gcpConfig.projects,
      configured: projects,
    },
    aws: {
      allowed: Object.keys(awsConfig.account),
      configured: accounts,
    },
    pulumi: {
      accessTokens: pulumis,
    },
    tailscale: {
      clients: tailscales,
    },
    doppler: {
      projects: Object.keys(dopplerEnvironments),
    },
    vault: {
      projects: vaultStores.apply((stores) => Object.keys(stores)),
    },
    repositories: Object.fromEntries(
      repositories.map((repository) => [
        repository.name,
        {
          google: repository.accessPermissions?.google?.project != undefined,
          gcs: getOrDefault(
            repository.accessPermissions?.google?.hmacKey,
            false,
          ),
          aws: repository.accessPermissions?.aws?.account != undefined,
          pulumi: getOrDefault(repository.accessPermissions?.pulumi, false),
          doppler: getOrDefault(repository.accessPermissions?.doppler, false),
          vault: getOrDefault(
            repository.accessPermissions?.vault?.enabled,
            true,
          ),
          tailscale: getOrDefault(
            repository.accessPermissions?.tailscale,
            false,
          ),
        },
      ]),
    ),
  };
};
