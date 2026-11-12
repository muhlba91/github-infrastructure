package google

import (
	"encoding/base64"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/muhlba91/github-infrastructure/pkg/lib/secrets"
	libVault "github.com/muhlba91/github-infrastructure/pkg/lib/vault"
	repoConf "github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/iam/role"
	gsa "github.com/muhlba91/pulumi-shared-library/pkg/lib/google/iam/serviceaccount"
	rModel "github.com/muhlba91/pulumi-shared-library/pkg/model/rotation"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/serviceaccount"
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
// rotated), written to the secret store as gcp-<account ID> (field "credentials", the key file's JSON).
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// stores: The secret stores (see secrets.ConfigureStores).
// providers: GCP providers by project.
func createServiceAccounts(ctx *pulumi.Context,
	project *google.RepositoryProject,
	stores map[string]*secrets.Store,
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

		if vErr := storeServiceAccountKey(ctx, project, name, account, key, stores); vErr != nil {
			log.Err(vErr).Msgf("[google][serviceaccount] error storing key of service account: %s", name)
			return vErr
		}
	}

	return nil
}

// storeServiceAccountKey writes a service account key to the secret store as gcp-<account ID>, field "credentials":
// the repository's store, or only the configured Vault mount.
// ctx: Pulumi context for resource management.
// project: The repository project configuration.
// name: The service account ID.
// account: The service account configuration.
// key: The service account key.
// stores: The secret stores (see secrets.ConfigureStores).
func storeServiceAccountKey(ctx *pulumi.Context,
	project *google.RepositoryProject,
	name string,
	account *google.RepositoryServiceAccount,
	key *serviceaccount.Key,
	stores map[string]*secrets.Store,
) error {
	store := stores[*project.Repository]
	if account.VaultMount != nil {
		store = stores[libVault.AdditionalMountKey(*account.VaultMount)]
		if store == nil {
			// not created by this program: it must exist already
			store = &secrets.Store{VaultPath: *account.VaultMount}
		}
	}

	credentials, _ := key.PrivateKey.ApplyT(func(privateKey string) (string, error) {
		data, err := base64.StdEncoding.DecodeString(privateKey)
		return string(data), err
	}).(pulumi.StringOutput)
	secretCredentials, _ := pulumi.ToSecret(credentials).(pulumi.StringOutput)

	return store.Write(ctx, fmt.Sprintf("gcp-%s", name), pulumi.StringMap{"credentials": secretCredentials}, key)
}
