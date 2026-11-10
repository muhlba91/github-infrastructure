package google

// RepositoryProject defines a Google project for a repository.
type RepositoryProject struct {
	// Repository is the name of the repository.
	Repository *string
	// Name is the name of the project.
	Name *string
	// Region is the region of the project.
	Region *string
	// IAMPermissions are the IAM permissions for the project.
	IAMPermissions []string
	// EnabledServices are the enabled services for the project.
	EnabledServices []string
	// LinkedProjects are the linked Google projects.
	LinkedProjects map[string]*RepositoryLinkedProject
	// HMACKey indicates if HMAC keys are enabled for the project.
	HMACKey *bool
	// ServiceAccounts are additional service accounts in the project, keyed by the account ID.
	ServiceAccounts map[string]*RepositoryServiceAccount
}

// RepositoryServiceAccount defines an additional service account of a repository.
type RepositoryServiceAccount struct {
	// Roles maps a project ID to the predefined roles the account gets in it.
	Roles map[string][]string
	// KeyRotationDays rotates the key every this many days (nil: never).
	KeyRotationDays *int
	// VaultMount is the KV mount the key is stored in (nil: the repository's mount).
	VaultMount *string
}

// RepositoryLinkedProject defines a linked Google project.
type RepositoryLinkedProject struct {
	// AccessLevel is the access level for the linked project.
	AccessLevel string
	// IAMPermissions are the IAM permissions for the linked project.
	IAMPermissions []string
}
