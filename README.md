# GitHub Infrastructure

[![Build status](https://img.shields.io/github/actions/workflow/status/muhlba91/github-infrastructure/pipeline.yml?style=for-the-badge)](https://github.com/muhlba91/github-infrastructure/actions/workflows/pipeline.yml)
[![License](https://img.shields.io/github/license/muhlba91/github-infrastructure?style=for-the-badge)](LICENSE.md)
[![](https://api.scorecard.dev/projects/github.com/muhlba91/github-infrastructure/badge?style=for-the-badge)](https://scorecard.dev/viewer/?uri=github.com/muhlba91/github-infrastructure)

This repository contains the automation for [GitHub Repositories](https://github.com) with optional Cloud Access using [Pulumi](http://pulumi.com).

---

## Requirements

- [Go](https://golang.org/dl/)
- [Pulumi](https://www.pulumi.com/docs/install/)

## Creating the Infrastructure

To create the services, a [Pulumi Stack](https://www.pulumi.com/docs/concepts/stack/) with the correct configuration needs to exists.

The stack can be deployed via:

```bash
pulumi up
```

## Destroying the Infrastructure

The entire infrastructure can be destroyed via:

```bash
pulumi destroy
```

**Attention**: you must set `ALLOW_REPOSITORY_DELETION="true"` as an environment variable to be able to delete repositories!

## Environment Variables

To successfully run, and configure the Pulumi plugins, you need to set a list of environment variables. Alternatively, refer to the used Pulumi provider's configuration documentation.

- `ALLOW_REPOSITORY_DELETION`: set to `true` to allow repository deletion
- `IGNORE_UNMANAGED_REPOSITORIES`: set to `true` to skip repositories not defined in `assets/repositories/`
- `AWS_REGION`: the AWS region to use
- `AWS_ACCESS_KEY_ID`: the AWS secret key
- `AWS_SECRET_ACCESS_KEY`: the AWS secret access key
- `CLOUDSDK_COMPUTE_REGION`: the Google Cloud (GCP) region
- `GOOGLE_APPLICATION_CREDENTIALS`: reference to a file containing the Google Cloud (GCP) service account credentials
- `SCW_ACCESS_KEY`: the Scaleway access key
- `SCW_SECRET_KEY`: the Scaleway secret key
- `SCW_ORGANIZATION_ID`: the Scaleway organization ID
- `SCW_PROJECT_ID`: the Scaleway project ID
- `SCW_DEFAULT_REGION`: the Scaleway default region
- `SCW_DEFAULT_ZONE`: the Scaleway default zone
- `GITHUB_TOKEN`: the GitHub token with permissions to manage repositories
- `GITLAB_TOKEN`: the GitLab token with permissions to manage access
- `PULUMI_ACCESS_TOKEN`: the Pulumi access token
- `OAUTH_CLIENT_ID`: Tailscale OAuth client ID
- `OAUTH_CLIENT_SECRET`: Tailscale OAuth client secret

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

### Scaleway

Scaleway configuration is based on each allowed project.

```yaml
scaleway:
  defaultRegion: the default region for every project
  defaultZone: the default zone for every project
  organizationID: the Scaleway organization ID
  projects: a map containing all allowed project identifiers
```

### Vault

Vault connection configuration. The token will be retrieved from the corresponding stack's output.

Attention: Vault will only be used if a connection configuration can be created.

```yaml
vault:
  address: the URL to the Vault instance
  enabled: whether Vault integration is enabled
```

#### Repository YAML

Repositories are defined in YAML format. For each repository to create a YAML file must be created in [assets/repositories/](assets/repositories/).

The format is described in the [template](assets/templates/repository.yml).

---

## Continuous Integration and Automations

- [GitHub Actions](https://docs.github.com/en/actions) are linting, and verifying the code.
- [Renovate Bot](https://github.com/renovatebot/renovate) is updating Go modules, and GitHub Actions.
