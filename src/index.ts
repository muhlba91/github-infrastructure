import { configureAwsAccounts } from './lib/aws';
import { awsConfig, gcpConfig, repositories } from './lib/configuration';
import { createRepositories } from './lib/github';
import { configureGoogleProjects } from './lib/google';
import { configureTailscale } from './lib/tailscale';
import { getOrDefault } from './lib/util/get_or_default';
import { configureVaultStores } from './lib/vault';

export = async () => {
  const githubRepositories = createRepositories();

  const vaultStores = configureVaultStores(githubRepositories);
  const tailscales = vaultStores.apply((stores) => configureTailscale(stores));
  const projects = vaultStores.apply((stores) =>
    configureGoogleProjects(stores),
  );
  const accounts = vaultStores.apply((stores) => configureAwsAccounts(stores));

  return {
    google: {
      allowed: gcpConfig.projects,
      configured: projects,
    },
    aws: {
      allowed: Object.keys(awsConfig.account),
      configured: accounts,
    },
    tailscale: {
      clients: tailscales,
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
          vault:
            getOrDefault(repository.manageLifecycle, true) &&
            getOrDefault(repository.accessPermissions?.vault?.enabled, true),
          tailscale: getOrDefault(
            repository.accessPermissions?.tailscale,
            false,
          ),
        },
      ]),
    ),
  };
};
