import * as gcp from '@pulumi/gcp';
import { Resource } from '@pulumi/pulumi';
import * as doppler from '@pulumiverse/doppler';

import { GoogleRepositoryProjectData } from '../../model/data/google';
import { StringMap } from '../../model/map';
import { writeToDoppler } from '../util/doppler/secret';

/**
 * Creates IAM for a Google project.
 *
 * @param {GoogleRepositoryProjectData} project the Google project
 * @param {gcp.serviceaccount.Account} serviceAccount the service account for the project
 * @param {StringMap<gcp.Provider>} providers the providers for all projects
 * @param {StringMap<doppler.Environment>} dopplerEnvironments the doppler environments
 * @param {Resource[]} dependencies the Pulumi dependencies
 */
export const createHmacKey = (
  project: GoogleRepositoryProjectData,
  serviceAccount: gcp.serviceaccount.Account,
  providers: StringMap<gcp.Provider>,
  dopplerEnvironments: StringMap<doppler.Environment>,
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

  writeToDoppler(
    'GCS_ACCESS_KEY_ID',
    key.accessId,
    dopplerEnvironments[project.repository].project,
  );
  writeToDoppler(
    'GCS_SECRET_ACCESS_KEY',
    key.secret,
    dopplerEnvironments[project.repository].project,
  );
};
