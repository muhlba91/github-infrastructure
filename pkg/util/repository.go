package util //nolint:revive // package name is util

import (
	"os"
	"path/filepath"

	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"gopkg.in/yaml.v3"
)

// ParseRepositoriesFromFiles reads the repository configuration files from the specified directory.
func ParseRepositoriesFromFiles(dir string) ([]*repository.Config, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var repos []*repository.Config
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		b, rErr := os.ReadFile(full)
		if rErr != nil {
			return nil, rErr
		}

		var r repository.Config
		if yErr := yaml.Unmarshal(b, &r); yErr != nil {
			return nil, yErr
		}
		repos = append(repos, &r)
	}

	return repos, nil
}
