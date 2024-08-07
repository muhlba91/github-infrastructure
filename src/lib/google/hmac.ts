import * as gcp from '@pulumi/gcp';
import { all, Resource } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

import { GoogleRepositoryProjectData } from '../../model/data/google';
import { StringMap } from '../../model/map';
import { writeToVault } from '../util/vault/secret';
import { vaultProvider } from '../vault';

/**
 * Creates IAM for a Google project.
 *
 * @param {GoogleRepositoryProjectData} project the Google project
 * @param {gcp.serviceaccount.Account} serviceAccount the service account for the project
 * @param {StringMap<gcp.Provider>} providers the providers for all projects
 * @param {StringMap<vault.Mount>} vaultStores the vault stores
 * @param {Resource[]} dependencies the Pulumi dependencies
 */
export const createHmacKey = (
  project: GoogleRepositoryProjectData,
  serviceAccount: gcp.serviceaccount.Account,
  providers: StringMap<gcp.Provider>,
  vaultStores: StringMap<vault.Mount>,
  dependencies: Resource[],
) => {
  const key = new gcp.storage.HmacKey(
    `gcp-hmac-${project.repository}-${project.name}`,
    {
      serviceAccountEmail: serviceAccount.email,
      project: project.name,
    },
    {
      provider: providers[project.name],
      dependsOn: dependencies,
    },
  );

  writeToVault(
    'google-cloud-storage',
    all([key.accessId, key.secret]).apply(([accessId, secret]) =>
      JSON.stringify({
        access_key_id: accessId,
        secret_access_key: secret,
      }),
    ),
    vaultProvider,
    vaultStores[project.repository],
  );
};
