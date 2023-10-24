# GitHub Infrastructure

[![Build status](https://img.shields.io/github/actions/workflow/status/muhlba91/github-infrastructure/pipeline.yml?style=for-the-badge)](https://github.com/muhlba91/github-infrastructure/actions/workflows/pipeline.yml)
[![License](https://img.shields.io/github/license/muhlba91/github-infrastructure?style=for-the-badge)](LICENSE.md)

This repository contains the automation for [GitHub Repositories](https://github.com) with optional Cloud Access using [Pulumi](http://pulumi.com).

---

## Requirements

- [NodeJS](https://nodejs.org/en), and [yarn](https://yarnpkg.com)
- [Pulumi](https://www.pulumi.com/docs/install/)

## Creating the Infrastructure

To create the repositories, a [Pulumi Stack](https://www.pulumi.com/docs/concepts/stack/) with the correct configuration needs to exists.

The stack can be deployed via:

```bash
yarn install
yarn build; pulumi up
```

## Destroying the Infrastructure

The entire infrastructure can be destroyed via:

```bash
yarn install
yarn build; pulumi destroy
```

**Attention**: you must set `ALLOW_REPOSITORY_DELETION="true"` as an environment variable to be able to delete repositories!

## Environment Variables

To successfully run, and configure the Pulumi plugins, you need to set a list of environment variables. Alternatively, refer to the used Pulumi provider's configuration documentation.

- `AWS_REGION`: the AWS region to use
- `AWS_ACCESS_KEY_ID`: the AWS secret key
- `AWS_SECRET_ACCESS_KEY`: the AWS secret access key
- `CLOUDSDK_COMPUTE_REGION` the Google Cloud (GCP) region
- `GOOGLE_APPLICATION_CREDENTIALS`: reference to a file containing the Google Cloud (GCP) service account credentials
- `DOPPLER_TOKEN`: the Doppler token with permissions to manage projects
- `GITHUB_TOKEN`: the GitHub token with permissions to manage repositories

---

## Configuration

The following section describes the configuration which must be set in the Pulumi Stack.

***Attention:*** do use [Secrets Encryption](https://www.pulumi.com/docs/concepts/secrets/#:~:text=Pulumi%20never%20sends%20authentication%20secrets,“secrets”%20for%20extra%20protection.) provided by Pulumi for secret values!

### AWS

AWS configuration is based on each allowed account.

```yaml
aws:
  defaultRegion: the default region for every account
  account: a map of AWS accounts to IAM role configuration
    <ACCOUNT_ID>:
      roleArn: the IAM role ARN to assume with correct permissions
      externalId: the the ExternalID property to assume the role
```

### Google Cloud

Google Cloud configuration is based on each allowed project.

```yaml
google:
  allowHmacKeys: allows creating HMAC Google Cloud Storage keys
  defaultRegion: the default region for every project
  projects: a list containing all allowed project identifiers
```

### Repositories

Repositories configuration sets default values and GitHub account information.

```yaml
repositories:
  owner: the owner/organization of all repositories
  subscription: the subscription type of the user/organization (e.g. "none")
```

#### Repository YAML

Repositories are defined in YAML format. For each repository to create a YAML file must be created in [assets/repositories/](assets/repositories/).

The format is described in the [template](assets/templates/repository.yml).

---

## Continuous Integration and Automations

- [GitHub Actions](https://docs.github.com/en/actions) are linting, and verifying the code.
- [Renovate Bot](https://github.com/renovatebot/renovate) is updating NodeJS packages, and GitHub Actions.
