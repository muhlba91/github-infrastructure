import * as doppler from '@pulumiverse/doppler';

import { StringMap } from '../../model/map';
import { repositories } from '../configuration';

import { createProject } from './project';

/**
 * Creates all Doppler projects.
 */
export const configureDoppler = (): StringMap<doppler.Environment> => {
  return Object.fromEntries(
    repositories.map((repository) => [
      repository.name,
      createProject(repository),
    ]),
  );
};
