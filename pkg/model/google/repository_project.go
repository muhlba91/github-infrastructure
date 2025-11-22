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
}

// RepositoryLinkedProject defines a linked Google project.
type RepositoryLinkedProject struct {
	// AccessLevel is the access level for the linked project.
	AccessLevel string
	// IAMPermissions are the IAM permissions for the linked project.
	IAMPermissions []string
}
