package aws

import (
	"fmt"

	awsConf "github.com/muhlba91/github-infrastructure/pkg/model/config/aws"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
)

const (
	// effectAllow is the effect of the trust statements.
	effectAllow = "Allow"
	// keyEffect is the key of the effect in a trust statement.
	keyEffect = "Effect"
	// keyAction is the key of the action in a trust statement.
	keyAction = "Action"
)

// trustPolicy builds the assume role policy of a CI role: GitHub OIDC for the repository and, if configured, the
// IAM Identity Center permission set (local runs). Without an organization ID the permission set is not trusted.
// identityProviderArn: ARN of the AWS IAM Identity Provider for GitHub OIDC.
// owner: Owner of the GitHub repository.
// repository: Name of the GitHub repository.
// sso: The IAM Identity Center configuration, nil if the permission set must not be trusted.
func trustPolicy(identityProviderArn, owner, repository string, sso *awsConf.SSO) map[string]any {
	//nolint:gosec // false positive, this is not a hardcoded secret but a condition for the OIDC token
	statements := []map[string]any{
		{
			keyEffect: effectAllow,
			keyAction: "sts:AssumeRoleWithWebIdentity",
			"Principal": map[string]any{
				"Federated": identityProviderArn,
			},
			"Condition": map[string]any{
				"StringEquals": map[string]any{
					"token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
				},
				"StringLike": map[string]any{
					"token.actions.githubusercontent.com:sub": fmt.Sprintf("repo:%s/%s:*", owner, repository),
				},
			},
		},
	}

	if sso != nil && defaults.GetOrDefault(sso.OrganizationID, "") != "" {
		statements = append(statements, ssoStatement(*sso.OrganizationID, sso.PermissionSetName))
	}

	return map[string]any{
		"Version":   "2012-10-17",
		"Statement": statements,
	}
}

// ssoStatement builds the trust statement for the IAM Identity Center permission set. The role name carries a
// hash per account and the path the home region, so the principal is matched by pattern within the organization.
// organizationID: ID of the AWS Organization.
// permissionSetName: Name of the permission set, defaults to ci-assume.
func ssoStatement(organizationID string, permissionSetName *string) map[string]any {
	name := defaults.GetOrDefault(permissionSetName, defaultSSOPermissionSet)

	return map[string]any{
		keyEffect: effectAllow,
		keyAction: "sts:AssumeRole",
		"Principal": map[string]any{
			"AWS": "*",
		},
		"Condition": map[string]any{
			"StringEquals": map[string]any{
				"aws:PrincipalOrgID": organizationID,
			},
			"ArnLike": map[string]any{
				"aws:PrincipalArn": fmt.Sprintf(
					"arn:aws:iam::*:role/aws-reserved/sso.amazonaws.com/*/AWSReservedSSO_%s_*",
					name,
				),
			},
		},
	}
}
