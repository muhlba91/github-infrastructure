package scaleway

// RepositoryProject defines a Scaleway project for a repository.
type RepositoryProject struct {
	// Repository is the name of the repository.
	Repository *string
	// Name is the name of the project.
	Name *string
	// OrganizationID is the ID of the Scaleway organization.
	OrganizationID *string
	// Region is the region of the project.
	Region *string
	// Zone is the zone of the project.
	Zone *string
	// IAMPermissions are the IAM permissions for the project.
	IAMPermissions []string
	// LinkedProjects are the linked Scaleway projects.
	LinkedProjects map[string]*RepositoryLinkedProject
}

// RepositoryLinkedProject defines a linked Scaleway project.
type RepositoryLinkedProject struct {
	// AccessLevel is the access level for the linked project.
	AccessLevel string
	// IAMPermissions are the IAM permissions for the linked project.
	IAMPermissions []string
}
