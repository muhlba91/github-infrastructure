/**
 * Defines Tailscale config.
 */
export interface TailscaleConfig {
  readonly oauth: TailscaleOAuthConfig;
}

/**
 * Defines Tailscale OAuth config.
 */
export interface TailscaleOAuthConfig {
  readonly id: string;
  readonly secret: string;
}
