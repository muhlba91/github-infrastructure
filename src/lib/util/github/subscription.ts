import { repositoriesConfig } from '../../configuration';

/**
 * Checks if the user has a subscription.
 *
 * @returns {boolean} true if the user has a subscription; false otherwise
 */
export const hasSubscription = (): boolean =>
  repositoriesConfig.subscription != undefined &&
  repositoriesConfig.subscription != 'none';
