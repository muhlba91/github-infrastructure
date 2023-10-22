import * as doppler from '@pulumiverse/doppler';

import { RepositoryConfig } from '../../model/config/repository';
import { repositoriesConfig } from '../configuration';

/**
 * Creates a Doppler project.
 *
 * @param {RepositoryConfig} repository the repository
 * @returns {doppler.Environment} the environment
 */
export const createProject = (
  repository: RepositoryConfig,
): doppler.Environment => {
  const project = new doppler.Project(
    `doppler-project-${repository.name}`,
    {
      name: repository.name,
      description: `GitHub repository: ${repositoriesConfig.owner}/${repository.name}`,
    },
    {},
  );

  const dopplerEnvironment = new doppler.Environment(
    `doppler-environment-prod-${repository.name}`,
    {
      project: project.name,
      slug: 'prod',
      name: 'Production',
    },
    {},
  );

  console.log(
    `[doppler][manual-action][${repository.name}] link the Doppler project to the GitHub repository`,
  );

  return dopplerEnvironment;
};
