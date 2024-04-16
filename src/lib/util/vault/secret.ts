import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

/**
 * Stores a value in Vault.
 *
 * @param {string} key the key
 * @param {Output<string>} value the value
 * @param {vault.Mount} vaultStore the optional vault store
 * @param {vault.Provider} provider the Vault provider
 */
export const writeToVault = (
  key: string,
  value: Output<string>,
  provider: vault.Provider,
  vaultStore?: vault.Mount,
) => {
  vaultStore?.path?.apply(
    (storePath) =>
      new vault.kv.SecretV2(
        `vault-secret-${storePath}-${key}`,
        {
          mount: storePath,
          name: key,
          dataJson: value,
        },
        {
          dependsOn: [vaultStore],
          provider: provider,
        },
      ),
  );
};
