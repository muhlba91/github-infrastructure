import * as gcp from '@pulumi/gcp';
import { interpolate, Output, Resource } from '@pulumi/pulumi';

import {
  GoogleRepositoryProjectData,
  GoogleWorkloadIdentityPoolData,
} from '../../model/data/google';
import { StringMap } from '../../model/map';
import { repositoriesConfig } from '../configuration';
import { writeToDoppler } from '../util/doppler/secret';
import { createRandomString } from '../util/random';

/**
 * Creates IAM for a Google project.
 *
 * @param {GoogleRepositoryProjectData} project the Google project
 * @param {StringMap<gcp.Provider>} providers the providers for all projects
 * @param {GoogleWorkloadIdentityPoolData} workloadIdentityPool the workload identity pool
 * @param {Resource[]} dependencies the Pulumi dependencies
 * @returns {gcp.serviceaccount.Account} the created service account
 */
export const createProjectIam = (
  project: GoogleRepositoryProjectData,
  providers: StringMap<gcp.Provider>,
  workloadIdentityPool: GoogleWorkloadIdentityPoolData,
  dependencies: Resource[],
): gcp.serviceaccount.Account => {
  const ciPostfix = createRandomString(
    `gcp-iam-role-ci-${project.repository}-${project.name}`,
    {},
  ).result.apply((id) => id.toLowerCase());
  const truncatedRepository = project.repository.substring(0, 18);

  const projects = [project.name].concat(project.linkedProjects ?? []);

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
          permissions: project.iamPermissions.map((permission) => permission),
          project: project.name,
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

  writeToDoppler(
    'GOOGLE_WORKLOAD_IDENTITY_PROVIDER',
    workloadIdentityPool.workloadIdentityProvider.name,
    project.repository,
  );
  writeToDoppler(
    'GOOGLE_WORKLOAD_IDENTITY_SERVICE_ACCOUNT',
    ciServiceAccount.email,
    project.repository,
  );
  writeToDoppler(
    'CLOUDSDK_COMPUTE_REGION',
    Output.create(project.region),
    project.repository,
  );

  return ciServiceAccount;
};
