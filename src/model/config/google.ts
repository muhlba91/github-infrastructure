/**
 * Defines Google config.
 */
export type GcpConfig = {
  readonly defaultRegion: string;
  readonly projects: readonly string[];
  readonly allowHmacKeys: boolean;
};
