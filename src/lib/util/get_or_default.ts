/**
 * Evaluates if the property is undefined and returns either it's value or the default.
 *
 * @param {T} value the value to evaluate
 * @param {T} defaultValue the default value
 * @returns {T} either `value` or `defaultValue` if it is undefined
 */
export const getOrDefault = <T>(value: T, defaultValue: T): T =>
  value == undefined ? defaultValue : value;
