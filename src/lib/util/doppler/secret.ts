import { Output } from '@pulumi/pulumi';
import * as doppler from '@pulumiverse/doppler';

import { environment } from '../../configuration';

/**
 * Stores a value in Doppler.
 *
 * @param {string} key the key
 * @param {Output<string>} value the value
 * @param {Output<string>} project the project name
 */
export const writeToDoppler = (
  key: string,
  value: Output<string>,
  project: Output<string>,
) => {
  project.apply(
    (dopplerProject) =>
      new doppler.Secret(`doppler-${dopplerProject}-${key}`, {
        name: key,
        value: value,
        project: dopplerProject,
        config: environment,
      }),
  );
};
