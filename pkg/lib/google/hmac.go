package google

import (
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// createHMACKey creates an HMAC key for the specified Google Cloud project and writes it to the secret store.
// ctx: Pulumi context for resource management.
// project: Google Cloud project details.
// serviceAccount: Service account associated with the HMAC key.
// store: The repository's secret store.
// provider: GCP provider for resource creation.
func createHMACKey(ctx *pulumi.Context,
	project *google.RepositoryProject,
	serviceAccount *serviceaccount.Account,
	store *secrets.Store,
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
		log.Err(err).Msgf("[google][hmac] error creating HMAC key for Google Cloud project: %s", *project.Name)
		return err
	}

	return store.Write(ctx, "google-cloud-storage", pulumi.StringMap{
		"access_key_id":     key.AccessId,
		"secret_access_key": key.Secret,
	})
}
