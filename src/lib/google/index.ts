import * as gcp from '@pulumi/gcp';
import { Resource } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { RepositoryConfig } from '../../model/config/repository';
import {
  GoogleRepositoryProjectData,
  GoogleWorkloadIdentityPoolData,
} from '../../model/data/google';
import { StringMap } from '../../model/map';
import { gcpConfig, repositories } from '../configuration';
import { uniqueFilter } from '../util/filter';

import { createHmacKey } from './hmac';
import { createProjectIam } from './iam';
import { createProjectGitHubOidc } from './oidc';
import { enableProjectServices } from './services';

export const DEFAULT_PERMISSIONS = [
  'cloudkms.cryptoKeyVersions.useToDecrypt',
  'cloudkms.cryptoKeyVersions.useToEncrypt',
  'cloudkms.cryptoKeys.getIamPolicy',
  'cloudkms.cryptoKeys.setIamPolicy',
  'cloudkms.locations.get',
  'cloudkms.locations.list',
  'compute.regions.list',
  'iam.serviceAccountKeys.create',
  'iam.serviceAccountKeys.delete',
  'iam.serviceAccountKeys.disable',
  'iam.serviceAccountKeys.enable',
  'iam.serviceAccountKeys.get',
  'iam.serviceAccountKeys.list',
  'iam.serviceAccounts.create',
  'iam.serviceAccounts.delete',
  'iam.serviceAccounts.disable',
  'iam.serviceAccounts.enable',
  'iam.serviceAccounts.get',
  'iam.serviceAccounts.getIamPolicy',
  'iam.serviceAccounts.list',
  'iam.serviceAccounts.setIamPolicy',
  'iam.serviceAccounts.undelete',
  'iam.serviceAccounts.update',
  'resourcemanager.projects.get',
  'resourcemanager.projects.getIamPolicy',
  'resourcemanager.projects.setIamPolicy',
  'resourcemanager.projects.update',
  'storage.hmacKeys.create',
  'storage.hmacKeys.delete',
  'storage.hmacKeys.get',
  'storage.hmacKeys.list',
  'storage.hmacKeys.update',
  'storage.buckets.create',
  'storage.buckets.createTagBinding',
  'storage.buckets.delete',
  'storage.buckets.deleteTagBinding',
  'storage.buckets.get',
  'storage.buckets.getIamPolicy',
  'storage.buckets.getObjectInsights',
  'storage.buckets.list',
  'storage.buckets.listEffectiveTags',
  'storage.buckets.listTagBindings',
  'storage.buckets.setIamPolicy',
  'storage.buckets.update',
  'storage.multipartUploads.abort',
  'storage.multipartUploads.create',
  'storage.multipartUploads.list',
  'storage.multipartUploads.listParts',
  'storage.objects.create',
  'storage.objects.delete',
  'storage.objects.get',
  'storage.objects.getIamPolicy',
  'storage.objects.list',
  'storage.objects.setIamPolicy',
  'storage.objects.update',
];

const DEFAULT_SERVICES = [
  'iam.googleapis.com',
  'iamcredentials.googleapis.com',
  'cloudresourcemanager.googleapis.com',
  'cloudkms.googleapis.com',
  'storage.googleapis.com',
  'storage-component.googleapis.com',
  'compute.googleapis.com',
];

/**
 * Creates all Google related infrastructure.
 *
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @returns {StringMap<string[]>} the configured Google projects
 */
export const configureGoogleProjects = (
  vaultStores: StringMap<vault.Mount>,
): StringMap<string[]> => {
  const providers = Object.fromEntries(
    gcpConfig.projects.map((project) => [
      project,
      new gcp.Provider(`gcp-provider-${project}`, {
        project: project,
      }),
    ]),
  );

  const googleRepositoryProjects = repositories
    .filter((repo) => repo.accessPermissions?.google?.project)
    .filter(filterRepositoryByAllowedProjects)
    .map((repo) => ({
      repository: repo.name,
      name: repo.accessPermissions?.google?.project ?? '',
      region: repo.accessPermissions?.google?.region ?? gcpConfig.defaultRegion,
      iamPermissions: DEFAULT_PERMISSIONS.concat(
        repo.accessPermissions?.google?.iamPermissions ?? [],
      ),
      enabledServices: DEFAULT_SERVICES.concat(
        repo.accessPermissions?.google?.enabledServices ?? [],
      ),
      linkedProjects: repo.accessPermissions?.google?.linkedProjects,
      hmacKey: repo.accessPermissions?.google?.hmacKey,
    }));

  const enabledServices = gcpConfig.projects.flatMap((project) =>
    enableProjectServices(
      project,
      googleRepositoryProjects
        .filter((repositoryProject) =>
          filterGoogleProjectByProject(repositoryProject, project),
        )
        .flatMap((repositoryProject) =>
          repositoryProject.name == project ||
          (repositoryProject.linkedProjects ?? {})[project]?.accessLevel ==
            'full'
            ? repositoryProject.enabledServices
            : DEFAULT_SERVICES,
        )
        .filter(uniqueFilter),
      providers,
    ),
  );

  const workloadIdentityPools = Object.fromEntries(
    googleRepositoryProjects
      .map((repositoryProject) => repositoryProject.name)
      .filter(uniqueFilter)
      .map((repositoryProject) => [
        repositoryProject,
        createProjectGitHubOidc(
          repositoryProject,
          providers[repositoryProject],
          enabledServices,
        ),
      ]),
  );
  googleRepositoryProjects.forEach((repositoryProject) =>
    configureProject(
      repositoryProject,
      providers,
      workloadIdentityPools[repositoryProject.name],
      vaultStores,
      enabledServices,
    ),
  );

  return googleRepositoryProjects
    .flatMap((repostoryProject) =>
      [repostoryProject.name]
        .concat(Object.keys(repostoryProject.linkedProjects ?? {}))
        .map((project) => ({
          name: project,
          repository: repostoryProject.repository,
        })),
    )
    .reduce<StringMap<string[]>>((projects, project) => {
      const group = (projects[project.name] ?? []).concat(project.repository);
      return {
        ...projects,
        [project.name]: group,
      };
    }, {});
};

/**
 * Configures a Google project.
 *
 * @param {GoogleRepositoryProjectData} project the Google project
 * @param {StringMap<gcp.Provider>} providers the providers for all projects
 * @param {GoogleWorkloadIdentityPoolData} workloadIdentityPool the workload identity pool to assign permissions for
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @param {Resource[]} dependencies the Pulumi dependencies
 */
const configureProject = (
  project: GoogleRepositoryProjectData,
  providers: StringMap<gcp.Provider>,
  workloadIdentityPool: GoogleWorkloadIdentityPoolData,
  vaultStores: StringMap<vault.Mount>,
  dependencies: Resource[],
) => {
  const serviceAccount = createProjectIam(
    project,
    providers,
    workloadIdentityPool,
    vaultStores,
    dependencies,
  );
  if (gcpConfig.allowHmacKeys && project.hmacKey) {
    createHmacKey(
      project,
      serviceAccount,
      providers,
      vaultStores,
      dependencies,
    );
  }
};

/**
 * Filters the repository by the configured projects.
 *
 * @param {RepositoryConfig} repository the repository
 * @returns {boolean} true is all projects are configured; false otherwise
 */
const filterRepositoryByAllowedProjects = (
  repository: RepositoryConfig,
): boolean => {
  const mainProject = repository.accessPermissions?.google?.project;
  if (mainProject == undefined || !gcpConfig.projects.includes(mainProject)) {
    console.error(
      `[google][${repository.name}][${mainProject}] the repository references an unconfigured project`,
    );
    return false;
  }

  const linkedProjects = Object.keys(
    repository.accessPermissions?.google?.linkedProjects ?? {},
  ).every((project) => {
    if (project == undefined || !gcpConfig.projects.includes(project)) {
      console.error(
        `[google][${repository.name}][${project}] the repository references an unconfigured project`,
      );
      return false;
    }
    return true;
  });

  return linkedProjects == undefined ? true : linkedProjects;
};

/**
 * Filters the Google project by the given project.
 *
 * @param {RepositoryConfig} googleProject the repository
 * @param {string} project the project
 * @returns {boolean} true is all projects are configured; false otherwise
 */
const filterGoogleProjectByProject = (
  googleProject: GoogleRepositoryProjectData,
  project: string,
): boolean =>
  (googleProject.name == project ||
    Object.keys(googleProject.linkedProjects ?? {}).includes(project)) ??
  false;
