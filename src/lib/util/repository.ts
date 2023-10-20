import { readdirSync, readFileSync } from 'fs';

import { parse } from 'yaml';

import { RepositoryConfig } from '../../model/config/repository';

/**
 * Parses YAML files from a given path and converts them into a repository.
 *
 * @param {string} path the path
 * @returns {RepositoryConfig[]} the repositories
 */
export const parseRepositoriesFromFiles = (path: string): RepositoryConfig[] =>
  readdirSync(path)
    .map((file) => parse(readFileSync(`${path}/${file}`, 'utf8')))
    .map((content) => content as RepositoryConfig);
