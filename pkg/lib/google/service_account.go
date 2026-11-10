package google

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/muhlba91/github-infrastructure/pkg/lib/config"
	libVault "github.com/muhlba91/github-infrastructure/pkg/lib/vault"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/iam/role"
	gsa "github.com/muhlba91/pulumi-shared-library/pkg/lib/google/iam/serviceaccount"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	rModel "github.com/muhlba91/pulumi-shared-library/pkg/model/rotation"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-vault/sdk/v7/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// serviceAccountIDPattern is Google's rule for service account IDs: 6-30 characters, lowercase letters, digits and
// hyphens, starting with a letter and not ending with a hyphen.
var serviceAccountIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)

// repositoryServiceAccounts converts the configured service accounts of a repository.
// accounts: The service accounts from the repository configuration.
func repositoryServiceAccounts(
	accounts map[string]repoConf.GoogleServiceAccountConfig,
) map[string]*google.RepositoryServiceAccount {
	result := make(map[string]*google.RepositoryServiceAccount)
	for name, account := range accounts {
		var mount *string
		if account.Vault != nil {
			mount = account.Vault.Mount
		}
		result[name] = &google.RepositoryServiceAccount{
			Roles:           account.Roles,
			KeyRotationDays: account.KeyRotationDays,
			VaultMount:      mount,
		}
	}
	return result
}

// createServiceAccounts creates the additional service accounts of a repository in its default project (shared
// library: account, role members, key): the predefined roles in the configured projects and a key (optionally
// rotated), stored in Vault as gcp-<account ID> (field "credentials", the key file's JSON).
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// vaultStores: The Vault mounts (repositories by name, created additional mounts by libVault.AdditionalMountKey).
// providers: GCP providers by project.
func createServiceAccounts(ctx *pulumi.Context,
	project *google.RepositoryProject,
	vaultStores map[string]*vault.Mount,
	providers map[string]*gcp.Provider,
) error {
	for _, name := range slices.Sorted(maps.Keys(project.ServiceAccounts)) {
		account := project.ServiceAccounts[name]
		if !serviceAccountIDPattern.MatchString(name) {
			err := fmt.Errorf("invalid service account ID: %s", name)
			log.Err(err).
				Msgf("[google][serviceaccount] invalid service account for repository: %s", *project.Repository)
			return err
		}

		displayName := fmt.Sprintf("GitHub Repository: %s (%s)", *project.Repository, name)
		description := fmt.Sprintf(
			"Service account %s declared by the GitHub repository: %s",
			name,
			*project.Repository,
		)
		serviceAccount, _, saErr := gsa.CreateServiceAccount(ctx, &gsa.CreateOptions{
			Name:          name,
			DisplayName:   &displayName,
			Description:   &description,
			Project:       pulumi.String(*project.Name),
			PulumiOptions: []pulumi.ResourceOption{pulumi.Provider(providers[*project.Name])},
		})
		if saErr != nil {
			log.Err(saErr).Msgf("[google][serviceaccount] error creating service account: %s", name)
			return saErr
		}

		// the roles per project (each with its project's provider), not through CreateServiceAccount (one project)
		member := pulumi.Sprintf("serviceAccount:%s", serviceAccount.Email)
		for _, roleProject := range slices.Sorted(maps.Keys(account.Roles)) {
			_, mErr := role.CreateMember(ctx, fmt.Sprintf("%s-%s", name, roleProject), &role.MemberOptions{
				Member:  member,
				Roles:   account.Roles[roleProject],
				Project: pulumi.String(roleProject),
				PulumiOptions: []pulumi.ResourceOption{
					pulumi.Provider(providers[roleProject]),
					pulumi.DependsOn([]pulumi.Resource{serviceAccount}),
				},
			})
			if mErr != nil {
				log.Err(mErr).
					Msgf("[google][serviceaccount] error assigning roles in %s to service account: %s", roleProject, name)
				return mErr
			}
		}

		var keyRotation *rModel.Options
		if account.KeyRotationDays != nil {
			keyRotation = &rModel.Options{Days: *account.KeyRotationDays}
		}
		key, kErr := gsa.CreateKey(ctx, fmt.Sprintf("%s-%s", *project.Repository, name), &gsa.KeyOptions{
			ServiceAccount: serviceAccount.Name,
			Rotation:       keyRotation,
			PulumiOptions: []pulumi.ResourceOption{
				pulumi.Provider(providers[*project.Name]),
				pulumi.DependsOn([]pulumi.Resource{serviceAccount}),
			},
		})
		if kErr != nil {
			log.Err(kErr).Msgf("[google][serviceaccount] error creating key for service account: %s", name)
			return kErr
		}

		if vErr := storeServiceAccountKey(ctx, project, name, account, key, vaultStores); vErr != nil {
			log.Err(vErr).Msgf("[google][serviceaccount] error storing key of service account: %s", name)
			return vErr
		}
	}

	return nil
}

// storeServiceAccountKey stores a service account key in Vault as gcp-<account ID>, field "credentials".
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// name: The service account ID.
// account: The service account configuration.
// key: The service account key.
// vaultStores: The Vault mounts (repositories by name, created additional mounts by libVault.AdditionalMountKey).
func storeServiceAccountKey(ctx *pulumi.Context,
	project *google.RepositoryProject,
	name string,
	account *google.RepositoryServiceAccount,
	key *serviceaccount.Key,
	vaultStores map[string]*vault.Mount,
) error {
	path := fmt.Sprintf("github-%s", *project.Repository)
	mount := vaultStores[*project.Repository]
	if account.VaultMount != nil {
		path = *account.VaultMount
		// nil when the mount is not created by this program: it must exist already
		mount = vaultStores[libVault.AdditionalMountKey(path)]
	}

	dependsOn := []pulumi.Resource{key}
	if mount != nil {
		dependsOn = append(dependsOn, mount)
	}

	value, _ := key.PrivateKey.ApplyT(func(privateKey string) (string, error) {
		credentials, err := base64.StdEncoding.DecodeString(privateKey)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(map[string]string{"credentials": string(credentials)})
		return string(data), err
	}).(pulumi.StringOutput)

	secretValue, _ := pulumi.ToSecret(value).(pulumi.StringOutput)
	_, err := secret.Create(ctx, &secret.CreateOptions{
		Path:  path,
		Key:   fmt.Sprintf("gcp-%s", name),
		Value: secretValue,
		PulumiOptions: []pulumi.ResourceOption{
			pulumi.Provider(config.VaultProvider),
			pulumi.DependsOn(dependsOn),
		},
	})
	return err
}
