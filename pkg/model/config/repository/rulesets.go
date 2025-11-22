package repository

// RulesetsConfig defines repository rulesets config.
type RulesetsConfig struct {
	// Branch defines the repository branch protections config.
	Branch *RulesetConfig `yaml:"branch,omitempty"`
	// Tag defines the repository tag protections config.
	Tag *RulesetConfig `yaml:"tag,omitempty"`
}
