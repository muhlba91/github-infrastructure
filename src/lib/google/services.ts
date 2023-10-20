import * as gcp from '@pulumi/gcp';

import { StringMap } from '../../model/map';

/**
 * Enables Google services for a project.
 *
 * @param {string} project the Google project
 * @param {string[]} services the services to enable
 * @param {StringMap<gcp.Provider>} providers the providers for all projects
 * @returns {gcp.projects.Service[]} the enabled services
 */
export const enableProjectServices = (
  project: string,
  services: readonly string[],
  providers: StringMap<gcp.Provider>,
): gcp.projects.Service[] =>
  services.map(
    (service) =>
      new gcp.projects.Service(
        `gcp-project-service-${project}-${service}`,
        {
          service: service,
          project: project,
        },
        {
          provider: providers[project],
        },
      ),
  );
