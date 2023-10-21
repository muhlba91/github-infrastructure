import { Rotating } from '@pulumiverse/time';

/**
 * Creates a rotating time.
 *
 * @param {string} name the string name
 * @param {number} days the days when it should be rotated (default: 30)
 * @returns {Rotating} the rotation
 */
export const createRotation = (
  name: string,
  { days = 30 }: { readonly days?: number },
): Rotating =>
  new Rotating(
    `rotation-${name}`,
    {
      rotationDays: days,
    },
    {},
  );
