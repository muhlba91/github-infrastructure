import { configureAwsAccounts } from './lib/aws';
import { awsConfig, gcpConfig, repositories } from './lib/configuration';
import { configureDoppler } from './lib/doppler';
import { createRepositories } from './lib/github';
import { configureGoogleProjects } from './lib/google';
import { configurePulumi } from './lib/pulumi';
import { getOrDefault } from './lib/util/get_or_default';

export = async () => {
  createRepositories();
  configureDoppler();

  const pulumis = configurePulumi();
  const projects = configureGoogleProjects();
  const accounts = configureAwsAccounts();

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
        },
      ]),
    ),
  };
};
