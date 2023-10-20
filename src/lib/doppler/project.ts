import * as doppler from '@pulumiverse/doppler';

import { RepositoryConfig } from '../../model/config/repository';
import { repositoriesConfig } from '../configuration';

/**
 * Creates a Doppler project.
 *
 * @param {RepositoryConfig} repository the repository
 */
export const createProject = (repository: RepositoryConfig) => {
  const project = new doppler.Project(
    `doppler-project-${repository.name}`,
    {
      name: repository.name,
      description: `GitHub repository: ${repositoriesConfig.owner}/${repository.name}`,
    },
    {},
  );

  new doppler.Environment(
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
};
