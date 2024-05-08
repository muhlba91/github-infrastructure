/**
 * Defines Google config.
 */
export interface GcpConfig {
  readonly defaultRegion: string;
  readonly projects: readonly string[];
  readonly allowHmacKeys: boolean;
}
