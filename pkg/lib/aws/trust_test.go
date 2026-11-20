package aws_test

import (
	"reflect"
	"testing"

	"github.com/muhlba91/github-infrastructure/pkg/lib/aws"
	awsConf "github.com/muhlba91/github-infrastructure/pkg/model/config/aws"
)

const (
	testProviderArn = "arn:aws:iam::111111111111:oidc-provider/token.actions.githubusercontent.com"
	testOrgID       = "o-example123"
)

func statements(t *testing.T, policy map[string]any) []map[string]any {
	t.Helper()
	result, ok := policy["Statement"].([]map[string]any)
	if !ok {
		t.Fatalf("Statement has the type %T, want []map[string]any", policy["Statement"])
	}
	return result
}

func TestTrustPolicy_GitHubOnly(t *testing.T) {
	tests := map[string]*awsConf.SSO{
		"no sso":                   nil,
		"sso without organization": {},
	}
	for name, sso := range tests {
		t.Run(name, func(t *testing.T) {
			result := statements(t, aws.TrustPolicy(testProviderArn, "owner", "repo", sso))
			if len(result) != 1 {
				t.Fatalf("got %d statements, want 1", len(result))
			}
			if result[0]["Action"] != "sts:AssumeRoleWithWebIdentity" {
				t.Errorf("Action = %v, want sts:AssumeRoleWithWebIdentity", result[0]["Action"])
			}
			want := map[string]any{"Federated": testProviderArn}
			if !reflect.DeepEqual(result[0]["Principal"], want) {
				t.Errorf("Principal = %v, want %v", result[0]["Principal"], want)
			}
		})
	}
}

func TestTrustPolicy_GitHubSubject(t *testing.T) {
	result := statements(t, aws.TrustPolicy(testProviderArn, "owner", "repo", nil))

	condition, ok := result[0]["Condition"].(map[string]any)
	if !ok {
		t.Fatalf("Condition has the type %T, want map[string]any", result[0]["Condition"])
	}
	want := map[string]any{"token.actions.githubusercontent.com:sub": "repo:owner/repo:*"}
	if !reflect.DeepEqual(condition["StringLike"], want) {
		t.Errorf("StringLike = %v, want %v", condition["StringLike"], want)
	}
}

func TestTrustPolicy_SSODefaultPermissionSet(t *testing.T) {
	orgID := testOrgID
	result := statements(t, aws.TrustPolicy(testProviderArn, "owner", "repo", &awsConf.SSO{OrganizationID: &orgID}))

	if len(result) != 2 {
		t.Fatalf("got %d statements, want 2", len(result))
	}
	if result[0]["Action"] != "sts:AssumeRoleWithWebIdentity" {
		t.Errorf("first statement Action = %v, want the GitHub OIDC action", result[0]["Action"])
	}
	if result[1]["Action"] != "sts:AssumeRole" {
		t.Errorf("second statement Action = %v, want sts:AssumeRole", result[1]["Action"])
	}
	if principal := (map[string]any{"AWS": "*"}); !reflect.DeepEqual(result[1]["Principal"], principal) {
		t.Errorf("Principal = %v, want %v", result[1]["Principal"], principal)
	}
	want := map[string]any{
		"StringEquals": map[string]any{"aws:PrincipalOrgID": testOrgID},
		"ArnLike": map[string]any{
			"aws:PrincipalArn": "arn:aws:iam::*:role/aws-reserved/sso.amazonaws.com/*/AWSReservedSSO_" +
				aws.DefaultSSOPermissionSet + "_*",
		},
	}
	if !reflect.DeepEqual(result[1]["Condition"], want) {
		t.Errorf("Condition = %v, want %v", result[1]["Condition"], want)
	}
}

func TestTrustPolicy_SSOCustomPermissionSet(t *testing.T) {
	orgID := testOrgID
	name := "custom-set"
	sso := &awsConf.SSO{OrganizationID: &orgID, PermissionSetName: &name}
	result := statements(t, aws.TrustPolicy(testProviderArn, "owner", "repo", sso))

	condition, ok := result[1]["Condition"].(map[string]any)
	if !ok {
		t.Fatalf("Condition has the type %T, want map[string]any", result[1]["Condition"])
	}
	want := map[string]any{
		"aws:PrincipalArn": "arn:aws:iam::*:role/aws-reserved/sso.amazonaws.com/*/AWSReservedSSO_custom-set_*",
	}
	if !reflect.DeepEqual(condition["ArnLike"], want) {
		t.Errorf("ArnLike = %v, want %v", condition["ArnLike"], want)
	}
}
