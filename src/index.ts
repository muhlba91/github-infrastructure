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

/**

pulumi stack change-secrets-provider "gcpkms://projects/root-muehlbachler/locations/europe/keyRings/infrastructure-encryption/cryptoKeys/infrastructure-encryption"

pulumi config set --path repositories.owner muhlba91
pulumi config set --path repositories.subscription none

---

aws:
  main:
    id: 454228071914
    external: 1e908bfa-7a2b-42c5-af25-6833c641cee5
    role: arn:aws:iam::454228071914:role/ci-configuration
  tuxnet:
    id: 126125163971
    external: f9905222-083d-46c5-8ee8-4a9312ac0397
    role: arn:aws:iam::126125163971:role/ci-configuration
  shared:
    id: 215622641987
    external: da858103-afe6-4995-b40f-a84fa36b6bc2
    role: arn:aws:iam::215622641987:role/ci-configuration


pulumi config set --path aws.defaultRegion eu-west-1 

pulumi config set --path aws.account.454228071914.externalId 1e908bfa-7a2b-42c5-af25-6833c641cee5 --secret
pulumi config set --path aws.account.454228071914.roleArn 'arn:aws:iam::454228071914:role/ci-configuration' --secret

pulumi config set --path aws.account.126125163971.externalId f9905222-083d-46c5-8ee8-4a9312ac0397 --secret
pulumi config set --path aws.account.126125163971.roleArn 'arn:aws:iam::126125163971:role/ci-configuration' --secret

pulumi config set --path aws.account.215622641987.externalId da858103-afe6-4995-b40f-a84fa36b6bc2 --secret
pulumi config set --path aws.account.215622641987.roleArn 'arn:aws:iam::215622641987:role/ci-configuration' --secret

---

google:
  root: root-muehlbachler
  tuxnet: tuxnet-385112
  mail: muehlbachler-mail-397612


pulumi config set --path 'google.defaultRegion' europe-west4
pulumi config set --path google.allowHmacKeys true

pulumi config set --path 'google.projects[0]' root-muehlbachler
pulumi config set --path 'google.projects[1]' tuxnet-385112
pulumi config set --path 'google.projects[2]' muehlbachler-mail-397612

---

muehlbachler-shared-services:
  aws (shared)
  google (root)
homelab-kubernetes-home-infrastructure:
  aws (tuxnet)
  google (tuxnet)
homelab-esphome-firmware:
  google (tuxnet)
muehlbachler-io-global-dns:
  google (root)
muehlbachler-xyz-global-dns:
  google (root)
homelab-buildkite-agents-home-infrastructure:
  google (tuxnet)
homelab-kubernetes-public-infrastructure:
  google (tuxnet)
muehlbachler-mail-aliases-infrastructure:
  google (mail) // TODO:

pulumi-gcp-aws-ignored-test:
  aws (tuxnet)
  google (tuxnet)

 */
