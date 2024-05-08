/**
 * Filters uniquie values.
 *
 * @param value the value
 * @param index the current index
 * @param self the array
 * @returns true if this object is the first occurence; false otherwise
 */
export const uniqueFilter = (value: unknown, index: number, self: unknown[]) =>
  self.indexOf(value) === index;
