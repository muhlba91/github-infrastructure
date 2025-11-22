package aws

import (
	"fmt"

	awsModel "github.com/muhlba91/github-infrastructure/pkg/model/aws"
	awsConf "github.com/muhlba91/github-infrastructure/pkg/model/config/aws"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// Configure sets up AWS resources based on the provided configuration.
// ctx: Pulumi context for resource management.
// repositories: List of repository configurations.
// vaultStores: Map of Vault mount configurations.
// awsConfig: AWS configuration details.
// repositoriesConfig: Repository configuration details.
func Configure(ctx *pulumi.Context,
	repositories []*repoConf.Config,
	vaultStores map[string]*vault.Mount,
	awsConfig *awsConf.Config,
	repositoriesConfig *repositories.Config,
) (map[string][]string, error) {
	providers := createProviders(ctx, awsConfig)

	awsRepositoryAccounts := createAWSRepositoryAccounts(repositories, awsConfig)

	identityProviderArns, ipErr := ConfigureIdentityProviders(
		ctx,
		awsRepositoryAccounts,
		providers,
	)
	if ipErr != nil {
		return nil, ipErr
	}

	accounts := make(map[string][]string)
	for _, repositoryAccount := range awsRepositoryAccounts {
		_ = configureAccount(
			ctx,
			repositoryAccount,
			identityProviderArns[*repositoryAccount.ID],
			vaultStores[*repositoryAccount.Repository],
			repositoriesConfig,
			providers[*repositoryAccount.ID],
		)

		accountRepositoryMapping, armOk := accounts[*repositoryAccount.ID]
		if !armOk {
			accountRepositoryMapping = []string{}
		}
		accounts[*repositoryAccount.ID] = append(accountRepositoryMapping, *repositoryAccount.Repository)
	}

	return accounts, nil
}

// createProviders initializes AWS providers for each account defined in the AWS configuration.
// ctx: Pulumi context for resource management.
// awsConfig: AWS configuration details.
func createProviders(ctx *pulumi.Context, awsConfig *awsConf.Config) map[string]*aws.Provider {
	providers := make(map[string]*aws.Provider)

	for account, config := range awsConfig.Account {
		provider, _ := aws.NewProvider(ctx, fmt.Sprintf("aws-provider-%s", account), &aws.ProviderArgs{
			AssumeRoles: &aws.ProviderAssumeRoleArray{
				&aws.ProviderAssumeRoleArgs{
					RoleArn:    pulumi.String(*config.RoleARN),
					ExternalId: pulumi.String(*config.ExternalID),
				},
			},
		})
		providers[account] = provider
	}

	return providers
}

// filterRepositoryByAllowedAccounts checks if the repository's specified AWS account is configured in the AWS settings.
// repoAccessPermissionsAws: AWS access configuration for the repository.
// awsConfig: AWS configuration details.
func filterRepositoryByAllowedAccounts(
	repoAccessPermissionsAws repoConf.AwsAccessConfig,
	awsConfig *awsConf.Config,
) bool {
	mainAccount := repoAccessPermissionsAws.Account
	if mainAccount == nil || awsConfig.Account[*mainAccount] == nil {
		log.Error().Msgf("[aws][%v] the repository references an unconfigured account", *mainAccount)
		return false
	}

	return true
}

// createAWSRepositoryAccounts constructs a mapping of repository names to their corresponding AWS account configurations.
// repositories: List of repository configurations.
// awsConfig: AWS configuration details.
func createAWSRepositoryAccounts(
	repositories []*repoConf.Config,
	awsConfig *awsConf.Config,
) map[string]*awsModel.RepositoryAccount {
	awsRepositoryAccounts := make(map[string]*awsModel.RepositoryAccount)
	for _, repository := range repositories {
		repoAccessPermissions := defaults.GetOrDefault(
			repository.AccessPermissions,
			repoConf.AccessPermissionsConfig{},
		)
		repoAccessPermissionsAws := defaults.GetOrDefault(
			repoAccessPermissions.Aws,
			repoConf.AwsAccessConfig{},
		)

		if repoAccessPermissionsAws.Account != nil && *repoAccessPermissionsAws.Account != "" &&
			filterRepositoryByAllowedAccounts(repoAccessPermissionsAws, awsConfig) {
			account := defaults.GetOrDefault(repoAccessPermissionsAws.Account, "")
			region := defaults.GetOrDefault(
				repoAccessPermissionsAws.Region,
				*awsConfig.DefaultRegion,
			)

			awsRepositoryAccounts[repository.Name] = &awsModel.RepositoryAccount{
				Repository: &repository.Name,
				ID:         &account,
				Region:     &region,
				IAMPermissions: append(
					repoAccessPermissionsAws.IAMPermissions,
					defaultPermissions...,
				),
			}
		}
	}

	return awsRepositoryAccounts
}
