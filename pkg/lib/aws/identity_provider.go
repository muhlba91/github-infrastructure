package aws

import (
	"fmt"
	"maps"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	awsModel "github.com/muhlba91/github-infrastructure/pkg/model/aws"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/metadata"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// ConfigureIdentityProviders sets up identity providers for AWS accounts associated with the repositories.
// ctx: Pulumi context for resource management.
// awsRepositoryAccounts: Mapping of repository names to their AWS account configurations.
// providers: Mapping of AWS providers for each account.
func ConfigureIdentityProviders(ctx *pulumi.Context,
	awsRepositoryAccounts map[string]*awsModel.RepositoryAccount,
	providers map[string]*aws.Provider,
) (map[string]*pulumi.StringOutput, error) {
	identityArns := make(map[string]*pulumi.StringOutput)

	uniqueAccounts := make(map[string]bool)
	for _, repositoryAccount := range awsRepositoryAccounts {
		uniqueAccounts[*repositoryAccount.ID] = true
	}
	for repositoryAccount := range maps.Keys(uniqueAccounts) {
		oidc, oErr := createAccountGitHubOidc(
			ctx,
			repositoryAccount,
			providers[repositoryAccount],
		)
		if oErr != nil {
			log.Err(oErr).
				Msgf("[aws][identity-provider] error creating AWS IAM Identity Provider for account: %s", repositoryAccount)
			return nil, oErr
		}
		identityArns[repositoryAccount] = oidc
	}

	return identityArns, nil
}

// createAccountGitHubOidc sets up an AWS IAM OpenID Connect Provider for GitHub Actions in the specified AWS account.
// ctx: Pulumi context for resource management.
// account: AWS account identifier.
// provider: AWS provider configured for the specified account.
func createAccountGitHubOidc(ctx *pulumi.Context,
	account string,
	provider *aws.Provider,
) (*pulumi.StringOutput, error) {
	tags := config.CommonLabels()
	tags["purpose"] = "github-actions"

	identityProvider, err := iam.NewOpenIdConnectProvider(
		ctx,
		fmt.Sprintf("aws-iam-identity-provider-%s", account),
		&iam.OpenIdConnectProviderArgs{
			Url:           pulumi.String("https://token.actions.githubusercontent.com"),
			ClientIdLists: pulumi.StringArray{pulumi.String("sts.amazonaws.com")},
			ThumbprintLists: pulumi.StringArray{
				pulumi.String("ffffffffffffffffffffffffffffffffffffffff"),
			},
			Tags: metadata.LabelsToStringMap(tags),
		},
		pulumi.Provider(provider),
	)
	if err != nil {
		log.Err(err).Msgf("[aws][identity-provider] error creating AWS IAM Identity Provider for account: %s", account)
		return nil, err
	}

	return &identityProvider.Arn, nil
}
