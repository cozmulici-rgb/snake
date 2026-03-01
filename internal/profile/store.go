package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"snake/internal/game"
)

func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return "snake_profile.json"
	}
	return filepath.Join(dir, "snake", "profile.json")
}

func Load(path string) (game.Profile, error) {
	var p game.Profile
	data, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	if len(data) == 0 {
		return p, nil
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, nil
}

func Save(path string, p game.Profile) error {
	if path == "" {
		return errors.New("empty profile path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
