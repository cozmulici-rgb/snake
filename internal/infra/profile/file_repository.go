package profile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"snake/internal/app/session"
)

type FileRepository struct {
	path string
}

func NewFileRepository(path string) *FileRepository {
	if path == "" {
		path = DefaultPath()
	}
	return &FileRepository{path: path}
}

func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return "snake_profile.json"
	}
	return filepath.Join(dir, "snake", "profile.json")
}

func (r *FileRepository) Load(_ context.Context) (session.Profile, error) {
	var p session.Profile
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
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

func (r *FileRepository) Save(_ context.Context, p session.Profile) error {
	if r.path == "" {
		return errors.New("empty profile path")
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}
