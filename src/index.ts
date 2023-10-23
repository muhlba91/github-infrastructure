import { configureAwsAccounts } from './lib/aws';
import { awsConfig, gcpConfig, repositories } from './lib/configuration';
import { configureDoppler } from './lib/doppler';
import { createRepositories } from './lib/github';
import { configureGoogleProjects } from './lib/google';
import { configurePulumi } from './lib/pulumi';
import { configureTailscale } from './lib/tailscale';
import { getOrDefault } from './lib/util/get_or_default';

export = async () => {
  createRepositories();

  const dopplerEnvironments = configureDoppler();
  const pulumis = configurePulumi(dopplerEnvironments);
  const tailscales = configureTailscale(dopplerEnvironments);
  const projects = configureGoogleProjects(dopplerEnvironments);
  const accounts = configureAwsAccounts(dopplerEnvironments);

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
          tailscale: getOrDefault(
            repository.accessPermissions?.tailscale,
            false,
          ),
        },
      ]),
    ),
  };
};
