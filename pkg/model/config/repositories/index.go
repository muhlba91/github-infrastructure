package repositories

// Config defines Repositories-related configuration.
type Config struct {
	// Owner contains configuration for the repository owner.
	Owner *string `yaml:"owner,omitempty"`
	// Subscription indicates the GitHub subscription status.
	Subscription *string `yaml:"subscription,omitempty"`
}
