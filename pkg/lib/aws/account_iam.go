package aws

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	awsModel "github.com/muhlba91/github-infrastructure/pkg/model/aws"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/metadata"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createAccountIAM creates AWS IAM roles for Continuous Integration for the specified repository account.
// ctx: Pulumi context for resource management.
// account: The repository account configuration.
// identityProviderArn: ARN of the AWS IAM Identity Provider for GitHub OIDC.
// vaultStore: Vault mount point for storing secrets.
// repositoriesConfig: Configuration for the GitHub repositories.
// provider: AWS provider configured for the specific account.
func createAccountIAM(ctx *pulumi.Context,
	account *awsModel.RepositoryAccount,
	identityProviderArn string,
	vaultStore *vault.Mount,
	repositoriesConfig *repositories.Config,
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
		provider,
	)
	if rErr != nil {
		return nil, rErr
	}

	pulumi.All(vaultStore.Path, role.Arn).ApplyT(func(all []any) error {
		path, _ := all[0].(string)
		roleArn, _ := all[1].(string)

		value, _ := json.Marshal(map[string]string{
			"identity_role_arn": roleArn,
			"region":            *account.Region,
		})

		_, err := secret.Create(ctx, &secret.CreateOptions{
			Path:  path,
			Key:   "aws",
			Value: pulumi.String(value),
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(config.VaultProvider),
			},
		})
		return err
	})

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
// provider: AWS provider configured for the specific account.
func createRole(ctx *pulumi.Context,
	account *awsModel.RepositoryAccount,
	identityProviderArn string,
	repositoriesConfig *repositories.Config,
	tags map[string]string,
	truncatedRepository string,
	ciPostfix pulumi.StringOutput,
	provider *aws.Provider,
) (*iam.Role, error) {
	roleDoc, _ := json.Marshal(map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect": "Allow",
				"Action": "sts:AssumeRoleWithWebIdentity",
				"Principal": map[string]any{
					"Federated": identityProviderArn,
				},
				"Condition": map[string]any{
					"StringEquals": map[string]any{
						"token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
					},
					"StringLike": map[string]any{
						"token.actions.githubusercontent.com:sub": fmt.Sprintf(
							"repo:%s/%s:*",
							*repositoriesConfig.Owner,
							*account.Repository,
						),
					},
				},
			},
		},
	})

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
		return nil, paErr
	}

	return role, nil
}
