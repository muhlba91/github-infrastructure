import { Output } from '@pulumi/pulumi';
import * as doppler from '@pulumiverse/doppler';

import { environment } from '../../configuration';

/**
 * Stores a value in Doppler.
 *
 * @param {string} key the key
 * @param {Output<string>} value the value
 * @param {doppler.Environment} dopplerEnvironment the optional doppler environment
 */
export const writeToDoppler = (
  key: string,
  value: Output<string>,
  dopplerEnvironment?: doppler.Environment,
) => {
  dopplerEnvironment?.project?.apply(
    (dopplerProject) =>
      new doppler.Secret(
        `doppler-${dopplerProject}-${key}`,
        {
          name: key,
          value: value,
          project: dopplerProject,
          config: environment,
        },
        {
          dependsOn: [dopplerEnvironment],
        },
      ),
  );
};
