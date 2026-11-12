package google

import "github.com/muhlba91/github-infrastructure/pkg/model/google"

// Exports of unexported identifiers for the external tests (package google_test).

// SecretManagerService exports secretManagerService.
const SecretManagerService = secretManagerService

// ProjectServices exports projectServices.
func ProjectServices(repoProject *google.RepositoryProject, project string) []string {
	return projectServices(repoProject, project)
}

// DefaultServices returns a copy of defaultServices.
func DefaultServices() []string {
	return append([]string{}, defaultServices...)
}
