package google

// Config defines Google Cloud-related configuration.
type Config struct {
	// DefaultRegion contains configuration for the default Google Cloud region.
	DefaultRegion *string `yaml:"defaultRegion,omitempty"`
	// Projects contains a list of Google Cloud projects.
	Projects []string `yaml:"projects,omitempty"`
	// AllowHMACKeys indicates whether HMAC keys are allowed.
	AllowHMACKeys *bool `yaml:"allowHmacKeys,omitempty"`
}
