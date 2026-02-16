package scaleway

// Config defines Scaleway-related configuration.
type Config struct {
	// OrganizationID contains the ID of the Scaleway organization to use.
	OrganizationID *string `yaml:"organizationID,omitempty"`
	// DefaultRegion contains configuration for the default Scaleway region.
	DefaultRegion *string `yaml:"defaultRegion,omitempty"`
	// DefaultZone contains configuration for the default Scaleway zone.
	DefaultZone *string `yaml:"defaultZone,omitempty"`
	// Projects contains a map of Scaleway project names to their IDs.
	Projects map[string]*string `yaml:"projects,omitempty"`
}
