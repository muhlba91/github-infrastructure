package util //nolint:revive // package name is util

import (
	"os"
	"path/filepath"

	"github.com/muhlba91/github-infrastructure/pkg/model/config/repository"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ParseRepositoriesFromFiles reads the repository configuration files from the specified directory.
func ParseRepositoriesFromFiles(dir string) ([]*repository.Config, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Err(err).Msgf("[repository] error reading repository configuration directory: %s", dir)
		return nil, err
	}

	var repos []*repository.Config
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		b, rErr := os.ReadFile(full)
		if rErr != nil {
			log.Err(rErr).Msgf("[repository] error reading repository configuration file: %s", full)
			return nil, rErr
		}

		var r repository.Config
		if yErr := yaml.Unmarshal(b, &r); yErr != nil {
			log.Err(yErr).Msgf("[repository] error parsing repository configuration file: %s", full)
			return nil, yErr
		}
		repos = append(repos, &r)
	}

	return repos, nil
}
