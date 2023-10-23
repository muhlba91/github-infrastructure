import * as doppler from '@pulumiverse/doppler';

import { StringMap } from '../../model/map';
import { repositories } from '../configuration';
import { getOrDefault } from '../util/get_or_default';

import { createProject } from './project';

/**
 * Creates all Doppler projects.
 */
export const configureDoppler = (): StringMap<doppler.Environment> => {
  return Object.fromEntries(
    repositories
      .filter((repository) =>
        getOrDefault(repository.accessPermissions?.doppler, true),
      )
      .map((repository) => [repository.name, createProject(repository)]),
  );
};
