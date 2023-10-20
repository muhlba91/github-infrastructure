import { repositories } from '../configuration';

import { createProject } from './project';

/**
 * Creates all Doppler projects.
 */
export const configureDoppler = () => {
  repositories.map(createProject);
};
