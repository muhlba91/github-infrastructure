package aws

import (
	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"
	awsModel "github.com/muhlba91/github-infrastructure/pkg/model/aws"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// configureAccount sets up AWS resources for a specific repository account.
// ctx: Pulumi context for resource management.
// account: AWS repository account configuration.
// identityProviderArn: ARN of the identity provider associated with the account.
// store: The repository's secret store.
// repositoriesConfig: Repository configuration details.
// awsConfig: AWS configuration details.
// provider: Pulumi AWS provider for resource creation.
func configureAccount(ctx *pulumi.Context,
	account *awsModel.RepositoryAccount,
	identityProviderArn *pulumi.StringOutput,
	store *secrets.Store,
	repositoriesConfig *repositories.Config,
	provider *aws.Provider,
) pulumi.Output {
	return identityProviderArn.ApplyT(func(arn string) error {
		_, err := createAccountIAM(
			ctx,
			account,
			arn,
			store,
			repositoriesConfig,
			provider,
		)
		if err != nil {
			log.Err(err).
				Msgf("[aws][account] error configuring AWS IAM for repository account: %s", *account.Repository)
			return err
		}

		return nil
	})
}
