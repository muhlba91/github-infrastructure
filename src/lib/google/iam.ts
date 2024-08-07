import * as gcp from '@pulumi/gcp';
import { all, interpolate, Resource } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import {
  GoogleRepositoryProjectData,
  GoogleWorkloadIdentityPoolData,
} from '../../model/data/google';
import { StringMap } from '../../model/map';
import { repositoriesConfig } from '../configuration';
import { createRandomString } from '../util/random';
import { writeToVault } from '../util/vault/secret';
import { vaultProvider } from '../vault';

import { DEFAULT_PERMISSIONS } from '.';

/**
 * Creates IAM for a Google project.
 *
 * @param {GoogleRepositoryProjectData} project the Google project
 * @param {StringMap<gcp.Provider>} providers the providers for all projects
 * @param {GoogleWorkloadIdentityPoolData} workloadIdentityPool the workload identity pool
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @param {Resource[]} dependencies the Pulumi dependencies
 * @returns {gcp.serviceaccount.Account} the created service account
 */
export const createProjectIam = (
  project: GoogleRepositoryProjectData,
  providers: StringMap<gcp.Provider>,
  workloadIdentityPool: GoogleWorkloadIdentityPoolData,
  vaultStores: StringMap<vault.Mount>,
  dependencies: Resource[],
): gcp.serviceaccount.Account => {
  const ciPostfix = createRandomString(
    `gcp-iam-role-ci-${project.repository}-${project.name}`,
    {},
  ).result.apply((id) => id.toLowerCase());
  const truncatedRepository = project.repository.substring(0, 18);

  const projects = [project.name].concat(
    Object.keys(project.linkedProjects ?? {}),
  );

  const ciRoles = Object.fromEntries(
    projects.map((name) => [
      name,
      new gcp.projects.IAMCustomRole(
        `gcp-iam-role-ci-${project.repository}-${name}`,
        {
          roleId: interpolate`ci.${truncatedRepository.replace(
            /-/g,
            '_',
          )}.${ciPostfix}`,
          title: `GitHub Repository: ${project.repository}`,
          description: `Continuous Integration role for the GitHub repository: ${project.repository}`,
          stage: 'GA',
          permissions:
            name == project.name ||
            (project.linkedProjects ?? {})[name]?.accessLevel == 'full'
              ? project.iamPermissions.map((permission) => permission)
              : DEFAULT_PERMISSIONS.concat(
                  (project.linkedProjects ?? {})[name].iamPermissions ?? [],
                ),
          project: name,
        },
        {
          provider: providers[name],
          dependsOn: dependencies,
        },
      ),
    ]),
  );

  const ciServiceAccount = new gcp.serviceaccount.Account(
    `gcp-iam-serviceaccount-ci-${project.repository}-${project.name}`,
    {
      accountId: interpolate`ci-${truncatedRepository}-${ciPostfix}`,
      displayName: `GitHub Repository: ${project.repository}`,
      description: `Continuous Integration Service Account for the GitHub repository: ${project.repository}`,
      project: project.name,
    },
    {
      provider: providers[project.name],
      dependsOn: dependencies,
    },
  );

  projects.forEach(
    (name) =>
      new gcp.projects.IAMMember(
        `gcp-iam-serviceaccount-ci-member-${project.repository}-${name}`,
        {
          member: interpolate`serviceAccount:${ciServiceAccount.email}`,
          role: ciRoles[name].id,
          project: name,
        },
        {
          provider: providers[name],
          dependsOn: dependencies,
        },
      ),
  );

  new gcp.serviceaccount.IAMBinding(
    `gcp-iam-identity-member-${project.repository}-${project.name}`,
    {
      serviceAccountId: ciServiceAccount.name,
      role: 'roles/iam.workloadIdentityUser',
      members: [
        interpolate`principalSet://iam.googleapis.com/${workloadIdentityPool.workloadIdentityPool.name}/attribute.repository/${repositoriesConfig.owner}/${project.repository}`,
      ],
    },
    {
      provider: providers[project.name],
      dependsOn: dependencies,
    },
  );

  writeToVault(
    'google-cloud',
    all([
      workloadIdentityPool.workloadIdentityProvider.name,
      ciServiceAccount.email,
    ]).apply(([workloadIdentityProviderName, serviceAccountEmail]) =>
      JSON.stringify({
        workload_identity_provider: workloadIdentityProviderName,
        ci_service_account: serviceAccountEmail,
        region: project.region,
      }),
    ),
    vaultProvider,
    vaultStores[project.repository],
  );

  return ciServiceAccount;
};
