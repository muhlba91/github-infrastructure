import { RepositoryConfig } from '../../../model/config/repository';

/**
 * Checks if the repository is private.
 *
 * @param {RepositoryConfig} config the repository config
 * @returns {boolean} true if the repository is private; false otherwise
 */
export const isPrivate = (config: RepositoryConfig): boolean =>
  config.visibility == 'private';
