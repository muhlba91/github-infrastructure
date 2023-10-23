/**
 * Defines Tailscale config.
 */
export type TailscaleConfig = {
  readonly oauth: TailscaleOAuthConfig;
};

/**
 * Defines Tailscale OAuth config.
 */
export type TailscaleOAuthConfig = {
  readonly id: string;
  readonly secret: string;
};
