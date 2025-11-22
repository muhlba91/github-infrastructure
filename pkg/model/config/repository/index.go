package repository

// Config defines Repository-related configuration.
type Config struct {
	// Name is the name of the repository.
	Name string `yaml:"name"`
	// Description is the description of the repository.
	Description string `yaml:"description"`
	// ManageLifecycle indicates whether to manage the repository lifecycle.
	ManageLifecycle *bool `yaml:"manageLifecycle,omitempty"`
	// Visibility defines the visibility of the repository.
	Visibility *string `yaml:"visibility,omitempty"`
	// Protected indicates whether the repository is protected.
	Protected *bool `yaml:"protected,omitempty"`
	// Topics are the topics associated with the repository.
	Topics []string `yaml:"topics,omitempty"`
	// Homepage is the homepage URL of the repository.
	Homepage *string `yaml:"homepage,omitempty"`
	// EnableWiki indicates whether to enable the wiki feature.
	EnableWiki *bool `yaml:"enableWiki,omitempty"`
	// EnableDiscussions indicates whether to enable the discussions feature.
	EnableDiscussions *bool `yaml:"enableDiscussions,omitempty"`
	// CreateProject indicates whether to create a project for the repository.
	CreateProject *bool `yaml:"createProject,omitempty"`
	// PagesBranch is the branch used for GitHub Pages.
	PagesBranch *string `yaml:"pagesBranch,omitempty"`
	// Rulesets defines the repository rulesets config.
	Rulesets *RulesetsConfig `yaml:"rulesets,omitempty"`
	// AccessPermissions defines the repository access permissions config.
	AccessPermissions *AccessPermissionsConfig `yaml:"accessPermissions,omitempty"`
}
