package aws

import (
	awsModel "github.com/muhlba91/github-infrastructure/pkg/model/aws"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// configureAccount sets up AWS resources for a specific repository account.
// ctx: Pulumi context for resource management.
// account: AWS repository account configuration.
// identityProviderArn: ARN of the identity provider associated with the account.
// vaultStore: Vault mount configuration for secrets management.
// repositoriesConfig: Repository configuration details.
// awsConfig: AWS configuration details.
// provider: Pulumi AWS provider for resource creation.
func configureAccount(ctx *pulumi.Context,
	account *awsModel.RepositoryAccount,
	identityProviderArn *pulumi.StringOutput,
	vaultStore *vault.Mount,
	repositoriesConfig *repositories.Config,
	provider *aws.Provider,
) pulumi.Output {
	return identityProviderArn.ApplyT(func(arn string) error {
		_, err := createAccountIAM(
			ctx,
			account,
			arn,
			vaultStore,
			repositoriesConfig,
			provider,
		)
		if err != nil {
			return err
		}

		return nil
	})
}
