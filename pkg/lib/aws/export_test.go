package aws

import awsConf "github.com/muhlba91/github-infrastructure/pkg/model/config/aws"

// Exports of unexported identifiers for the external tests (package aws_test).

// DefaultSSOPermissionSet exports defaultSSOPermissionSet.
const DefaultSSOPermissionSet = defaultSSOPermissionSet

// TrustPolicy exports trustPolicy.
func TrustPolicy(identityProviderArn, owner, repository string, sso *awsConf.SSO) map[string]any {
	return trustPolicy(identityProviderArn, owner, repository, sso)
}
