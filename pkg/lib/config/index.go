package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/rs/zerolog/log"

	"github.com/muhlba91/github-infrastructure/pkg/model/config/aws"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/google"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repositories"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/github-infrastructure/pkg/model/config/scaleway"
	vaultConf "github.com/muhlba91/github-infrastructure/pkg/model/config/vault"
	vaultData "github.com/muhlba91/github-infrastructure/pkg/model/vault"
	"github.com/muhlba91/github-infrastructure/pkg/util"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
)

//nolint:gochecknoglobals // global configuration is acceptable here
var (
	// Environment holds the current deployment environment (e.g., dev, staging, prod).
	Environment string
	// Stack holds the current Pulumi stack name.
	Stack *pulumi.StackReference
	// GlobalName is a constant name used across resources.
	GlobalName = "shared-services"
	// AllowRepositoryDeletion indicates whether repository deletion is permitted.
	AllowRepositoryDeletion = false
	// IgnoreUnmanagedRepositories indicates whether to ignore unmanaged repositories.
	IgnoreUnmanagedRepositories = false
	// HasVaultConnection indicates whether a Vault connection is configured.
	HasVaultConnection pulumi.BoolOutput
	// VaultConnectionConfig holds the Vault connection configuration.
	VaultConnectionConfig *vaultData.Config
	// VaultProvider holds the Vault provider resource.
	VaultProvider *vault.Provider
)

// LoadConfig loads the configuration for the given Pulumi context.
// ctx: The Pulumi context.
func LoadConfig(
	ctx *pulumi.Context,
) (*repositories.Config, *aws.Config, *google.Config, *scaleway.Config, *vaultConf.Config, []*repository.Config, error) {
	Environment = ctx.Stack()
	Stack, _ = pulumi.NewStackReference(
		ctx,
		fmt.Sprintf("%s/%s/%s", ctx.Organization(), ctx.Project(), Environment),
		nil,
	)

	cfg := config.New(ctx, "")

	var repositoriesConfig repositories.Config
	cfg.RequireObject("repositories", &repositoriesConfig)

	var awsConfig aws.Config
	cfg.RequireObject("aws", &awsConfig)

	var gcpConfig google.Config
	cfg.RequireObject("google", &gcpConfig)

	var scalewayConfig scaleway.Config
	cfg.RequireObject("scaleway", &scalewayConfig)

	var vaultConfig vaultConf.Config
	cfg.RequireObject("vault", &vaultConfig)

	repoDelEnv := strings.ToLower(os.Getenv("ALLOW_REPOSITORY_DELETION"))
	AllowRepositoryDeletion = defaults.GetOrDefault(&repoDelEnv, "false") == "true"

	unmanagedEnv := strings.ToLower(os.Getenv("IGNORE_UNMANAGED_REPOSITORIES"))
	IgnoreUnmanagedRepositories = defaults.GetOrDefault(&unmanagedEnv, "false") == "true"

	coreStack, sErr := pulumi.NewStackReference(
		ctx,
		fmt.Sprintf("%s/%s/%s", ctx.Organization(), "muehlbachler-core-infrastructure", Environment),
		nil,
	)
	if sErr != nil {
		log.Err(sErr).Msg("[config] error referencing core infrastructure stack for vault configuration")
		return nil, nil, nil, nil, nil, nil, sErr
	}
	cStackVault := coreStack.GetOutput(pulumi.String("vault"))
	HasVaultConnection = cStackVault.ApplyT(func(vaultConn any) bool { //nolint:errcheck // no error possible
		vKeys, _ := vaultConn.(map[string]any)["keys"].(map[string]any)
		vToken, _ := vKeys["rootToken"].(string)
		VaultConnectionConfig = &vaultData.Config{
			Address: vaultConfig.Address,
			Token:   &vToken,
		}

		VaultProvider, _ = vault.NewProvider(ctx, "vault", &vault.ProviderArgs{
			Address: pulumi.ToSecret(pulumi.StringPtr(*VaultConnectionConfig.Address)).(pulumi.StringPtrOutput),
			Token:   pulumi.ToSecret(pulumi.StringPtr(*VaultConnectionConfig.Token)).(pulumi.StringPtrOutput),
		})

		return *vaultConfig.Enabled && VaultConnectionConfig.Token != nil && *VaultConnectionConfig.Token != ""
	}).(pulumi.BoolOutput)

	repos, rErr := util.ParseRepositoriesFromFiles("./assets/repositories")
	if rErr != nil {
		log.Err(rErr).Msg("[config] error parsing repository configurations from files")
		return nil, nil, nil, nil, nil, nil, rErr
	}

	return &repositoriesConfig, &awsConfig, &gcpConfig, &scalewayConfig, &vaultConfig, repos, nil
}

// CommonLabels returns a map of common labels to be used across resources.
func CommonLabels() map[string]string {
	return map[string]string{
		"environment": Environment,
	}
}
