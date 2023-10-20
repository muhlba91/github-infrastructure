import { Output } from '@pulumi/pulumi';
import * as doppler from '@pulumiverse/doppler';

import { environment } from '../../configuration';

/**
 * Stores a value in Doppler.
 *
 * @param {string} key the key
 * @param {Output<string>} value the value
 * @param {string} project the project name
 */
export const writeToDoppler = (
  key: string,
  value: Output<string>,
  project: string,
) => {
  new doppler.Secret('doppler-' + project + '-' + key, {
    name: key,
    value: value,
    project: project,
    config: environment,
  });
};
