import { RandomString } from '@pulumi/random';

/**
 * Creates a random string.
 *
 * @param {string} name the string name
 * @param {number} length the length (default: 8)
 * @param {boolean} special enable special characters (default: false)
 * @returns {RandomString} the string
 */
export const createRandomString = (
  name: string,
  {
    length = 8,
    special = false,
  }: { readonly length?: number; readonly special?: boolean },
): RandomString =>
  new RandomString(
    `random-string-${name}`,
    {
      length: length,
      special: special,
      lower: true,
      upper: true,
      number: true,
    },
    {},
  );
