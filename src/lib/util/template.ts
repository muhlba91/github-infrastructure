import * as nunjucks from 'nunjucks';

/**
 * Reads and interpolates values in a template at given path.
 *
 * @param {string} path: the path to the template
 * @param {never} data: the values to interpolate
 * @returns {string} the rendered template
 */

export const renderTemplate = (path: string, data: object): string => {
  nunjucks.configure({ autoescape: false });
  return nunjucks.render(path, data);
};
