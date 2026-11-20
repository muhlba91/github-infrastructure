# Bootstrap (day 1)

This repository creates the GitHub repositories, Vault's mounts and every repository's CI identities, including its
own. Its inputs therefore never come from Vault, and it must be deployable from a developer machine (Mac) **without
CI**. This document lists everything that has to exist before the first `pulumi up`, and the order of the initial
bootstrap.

- **Day 1** is the initial bootstrap of everything, run from the Mac with your own logins.
- **Day 2** is everything after. CI works from day 2 on: its Google identity (workload identity) and its AWS identity (a role in the management account) are created by this
  repository's own first run (`assets/repositories/github-infrastructure.yml`).

## Prerequisites

Created by hand, outside of any repository:

1. **Google Cloud:** the projects listed in the stack configuration (`google.projects`), and in `public-services-402810`
   - the KMS key `europe/infrastructure-encryption/infrastructure-encryption` (Pulumi's secrets provider),
   - the Secret Manager API enabled (`gcloud services enable secretmanager.googleapis.com --project public-services-402810`;
     fnox needs it before the first run, afterwards this program manages it in all configured projects).
2. **Your Google login** on the Mac: `gcloud auth login` and `gcloud auth application-default login`, with the rights
   this program needs in all configured projects (owner).
3. **Pulumi Cloud:** the account and `pulumi login` on the Mac.
4. **GitHub:** a token managing the repositories (`GITHUB_TOKEN`).
5. **GitLab:** a token managing access (`GITLAB_TOKEN`).
6. **Tailscale:** an OAuth client (`OAUTH_CLIENT_ID`, `OAUTH_CLIENT_SECRET`).
7. **Scaleway:** an IAM application used only by this repository, with an API key and a policy granting `IAMManager` on
   the organization (the program creates the repositories' IAM applications, policies and API keys).
8. **AWS:** an Organization with IAM Identity Center.
   - A permission set `ci-assume` (session 1 h, inline policy: `sts:AssumeRole` on `arn:aws:iam::*:role/ci-*`), assigned to you on the management account.
   - In every account of the stack configuration (`aws.account`), a role (`ci-configuration`) that trusts the management account with an external ID and has the rights this program needs there.
   - Your login: `aws sso login --profile ci-assume` (the profile is in your `~/.aws/config`, see the dotfiles).
   - The Organization ID in the stack configuration (`aws.sso.organizationId`).
9. **mise** on the Mac (it installs Go, Pulumi and fnox from `mise.toml`).

## Secrets

The secrets live in Google Secret Manager in `public-services-402810`, one secret per environment variable, named
`github-infrastructure--<VARIABLE>`:

| Secret | Content |
| --- | --- |
| `github-infrastructure--GITHUB_TOKEN` | GitHub token |
| `github-infrastructure--GITLAB_TOKEN` | GitLab token |
| `github-infrastructure--OAUTH_CLIENT_ID` | Tailscale OAuth client: ID |
| `github-infrastructure--OAUTH_CLIENT_SECRET` | Tailscale OAuth client: secret |
| `github-infrastructure--SCW_ACCESS_KEY` | Scaleway API key: access key |
| `github-infrastructure--SCW_SECRET_KEY` | Scaleway API key: secret key |
| `github-infrastructure--AWS_IDENTITY_ROLE_ARN` | ARN of this repository's AWS CI role (not secret; seeded after the first run) |

Seed or rotate them with the script (it prompts for values without echo and never prints them):

```bash
bootstrap/seed-secrets.sh --check   # lists missing secrets
bootstrap/seed-secrets.sh           # creates missing secrets and prompts for values
```

Non-secret configuration (regions, Google project, Scaleway organization and project) is set as defaults in
`fnox.toml`; the remaining configuration is in the stack configuration.

## Order (day 1)

1. Prerequisites above.
2. The core infrastructure repository (`muehlbachler-core-infrastructure`, Vault) is deployed first: this program reads
   Vault's root token from its stack's outputs and writes the other repositories' secrets into Vault.
3. `bootstrap/seed-secrets.sh`, then `bootstrap/seed-secrets.sh --check`.
4. This repository, with the lease skipped because the AWS CI role does not exist yet (your Identity Center session assumes the roles in `aws.account` directly):

   ```bash
   # credentials of the ci-assume session for the AWS provider
   eval "$(aws configure export-credentials --profile ci-assume --format env)"
   fnox exec --except 'AWS_*' -- env AWS_REGION=eu-west-1 pulumi up --stack prod
   ```

   It also creates its own CI identities and writes the `GOOGLE_CLOUD_*` and `AWS_*` secrets into its GitHub repository.
5. Seed the role's ARN (`aws iam list-roles`, name `ci-github-infrastruct-<suffix>`) with `bootstrap/seed-secrets.sh`; the other secrets are kept.
6. Day 2: CI of this repository (and of all others) works, and `fnox exec -- pulumi preview --stack prod` assumes the CI role from your Identity Center login.

## Notes

- **Day 2 circularity:** this repository manages its own CI identity (service account, roles, workload identity binding,
  Secret Manager access, `GOOGLE_CLOUD_*` secrets). If a change breaks them, CI fails; repair it from the Mac with
  `fnox exec -- pulumi up --stack prod` (your own login, no CI needed).
- **AWS:** the CI role only assumes the roles in `aws.account`; their rights decide what this program can do. If the role or its trust breaks, repair it from the Mac with the day 1 command above.
- Missing rights of the CI identity are added to `accessPermissions.google.iamPermissions` in
  `assets/repositories/github-infrastructure.yml` and applied from the Mac.
- `mise.toml` sets `FNOX_PROFILE=default` in this repository, so a developer machine's `dev` fnox profile (which reads
  from Vault) is never used here.
- Local overrides go into `fnox.local.toml` (gitignored).
