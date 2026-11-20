package aws

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	awsModel "github.com/muhlba91/github-infrastructure/pkg/model/aws"
	awsConf "github.com/muhlba91/github-infrastructure/pkg/model/config/aws"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/metadata"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// createAccountIAM creates AWS IAM roles for Continuous Integration for the specified repository account.
// ctx: Pulumi context for resource management.
// account: The repository account configuration.
// identityProviderArn: ARN of the AWS IAM Identity Provider for GitHub OIDC.
// store: The repository's secret store.
// repositoriesConfig: Configuration for the GitHub repositories.
// sso: The IAM Identity Center configuration trusted by the role, nil for none.
// provider: AWS provider configured for the specific account.
func createAccountIAM(ctx *pulumi.Context,
	account *awsModel.RepositoryAccount,
	identityProviderArn string,
	store *secrets.Store,
	repositoriesConfig *repositories.Config,
	sso *awsConf.SSO,
	provider *aws.Provider,
) (*iam.Role, error) {
	tags := config.CommonLabels()
	tags["repository"] = *account.Repository
	tags["purpose"] = "github-repository"

	truncatedRepository := (*account.Repository)[:min(maxRepositoryLength, len(*account.Repository))]

	ciPostfix, ciPfErr := random.CreateString(
		ctx,
		fmt.Sprintf("random-string-aws-iam-role-ci-%s-%s", *account.Repository, *account.ID),
		&random.StringOptions{
			Length:  postfixLength,
			Special: false,
		},
	)
	if ciPfErr != nil {
		log.Err(ciPfErr).
			Msgf("[aws][iam] error creating random string for AWS IAM role postfix for repository: %s", *account.Repository)
		return nil, ciPfErr
	}

	role, rErr := createRole(
		ctx,
		account,
		identityProviderArn,
		repositoriesConfig,
		tags,
		truncatedRepository,
		ciPostfix.Text,
		sso,
		provider,
	)
	if rErr != nil {
		log.Err(rErr).Msgf("[aws][iam] error creating AWS IAM role for repository: %s", *account.Repository)
		return nil, rErr
	}

	stErr := store.Write(ctx, "aws", pulumi.StringMap{
		"identity_role_arn": role.Arn,
		"region":            pulumi.String(*account.Region),
	})
	if stErr != nil {
		log.Err(stErr).Msgf("[aws][iam] error storing AWS secrets for repository: %s", *account.Repository)
		return nil, stErr
	}

	return role, nil
}

// createRole creates an AWS IAM role for Continuous Integration for the specified repository account.
// ctx: Pulumi context for resource management.
// account: The repository account configuration.
// identityProviderArn: ARN of the AWS IAM Identity Provider for GitHub OIDC.
// repositoriesConfig: Configuration for the GitHub repositories.
// tags: Tags to be applied to the IAM role.
// truncatedRepository: Truncated name of the repository for naming purposes.
// ciPostfix: Random postfix for ensuring unique role names.
// sso: The IAM Identity Center configuration trusted by the role, nil for none.
// provider: AWS provider configured for the specific account.
func createRole(ctx *pulumi.Context,
	account *awsModel.RepositoryAccount,
	identityProviderArn string,
	repositoriesConfig *repositories.Config,
	tags map[string]string,
	truncatedRepository string,
	ciPostfix pulumi.StringOutput,
	sso *awsConf.SSO,
	provider *aws.Provider,
) (*iam.Role, error) {
	roleDoc, _ := json.Marshal(trustPolicy(identityProviderArn, *repositoriesConfig.Owner, *account.Repository, sso))

	//nolint:godox // TODO is required
	// FIXME: move to shared library
	role, rErr := iam.NewRole(
		ctx,
		fmt.Sprintf("aws-iam-role-ci-%s-%s", *account.Repository, *account.ID),
		&iam.RoleArgs{
			Name:             pulumi.Sprintf("ci-%s-%s", truncatedRepository, ciPostfix),
			Description:      pulumi.String(fmt.Sprintf("GitHub Repository: %s", *account.Repository)),
			AssumeRolePolicy: pulumi.String(roleDoc),
			Tags:             metadata.LabelsToStringMap(tags),
		},
		pulumi.Provider(provider),
	)
	if rErr != nil {
		log.Err(rErr).Msgf("[aws][iam] error creating AWS IAM role for repository: %s", *account.Repository)
		return nil, rErr
	}

	policyDoc, _ := json.Marshal(map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":   "Allow",
				"Action":   account.IAMPermissions,
				"Resource": "*",
			},
		},
	})
	//nolint:godox // TODO is required
	// FIXME: move to shared library
	policy, pErr := iam.NewPolicy(
		ctx,
		fmt.Sprintf("aws-iam-role-ci-policy-%s-%s", *account.Repository, *account.ID),
		&iam.PolicyArgs{
			Name:        pulumi.Sprintf("ci-%s-%s", truncatedRepository, ciPostfix),
			Description: pulumi.String(fmt.Sprintf("GitHub Repository: %s", *account.Repository)),
			Policy:      pulumi.String(policyDoc),
			Tags:        metadata.LabelsToStringMap(tags),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{
			role,
		}),
	)
	if pErr != nil {
		log.Err(pErr).Msgf("[aws][iam] error creating AWS IAM policy for repository: %s", *account.Repository)
		return nil, pErr
	}

	//nolint:godox // TODO is required
	// FIXME: move to shared library
	_, paErr := iam.NewRolePolicyAttachment(
		ctx,
		fmt.Sprintf("aws-iam-role-ci-policy-attachment-%s-%s", *account.Repository, *account.ID),
		&iam.RolePolicyAttachmentArgs{
			Role:      role.Name,
			PolicyArn: policy.Arn,
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{
			role,
			policy,
		}),
	)
	if paErr != nil {
		log.Err(paErr).Msgf("[aws][iam] error attaching AWS IAM policy to role for repository: %s", *account.Repository)
		return nil, paErr
	}

	return role, nil
}
