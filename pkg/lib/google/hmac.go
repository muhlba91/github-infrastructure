package google

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createHMACKey creates an HMAC key for the specified Google Cloud project and stores it in Vault.
// ctx: Pulumi context for resource management.
// project: Google Cloud project details.
// serviceAccount: Service account associated with the HMAC key.
// vaultStore: Vault mount where the HMAC key will be stored.
// provider: GCP provider for resource creation.
func createHMACKey(ctx *pulumi.Context,
	project *google.RepositoryProject,
	serviceAccount *serviceaccount.Account,
	vaultStore *vault.Mount,
	provider *gcp.Provider,
) error {
	key, err := storage.NewHmacKey(
		ctx,
		fmt.Sprintf("gcp-hmac-%s-%s", *project.Repository, *project.Name),
		&storage.HmacKeyArgs{
			ServiceAccountEmail: serviceAccount.Email,
			Project:             pulumi.String(*project.Name),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{serviceAccount}),
	)
	if err != nil {
		return err
	}

	pulumi.All(vaultStore.Path, key.AccessId, key.Secret).ApplyT(func(all []any) error {
		path, _ := all[0].(string)
		accessID, _ := all[1].(string)
		secretKey, _ := all[2].(string)

		value, _ := json.Marshal(map[string]string{
			"access_key_id":     accessID,
			"secret_access_key": secretKey,
		})

		_, vErr := secret.Write(ctx, &secret.WriteArgs{
			Path:  path,
			Key:   "google-cloud-storage",
			Value: pulumi.String(value),
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(config.VaultProvider),
			},
		})
		return vErr
	})

	return nil
}
