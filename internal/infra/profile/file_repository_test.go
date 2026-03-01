package profile

import (
	"context"
	"path/filepath"
	"testing"

	"snake/internal/app/session"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo := NewFileRepository(filepath.Join(dir, "profile.json"))

	in := session.Profile{
		BestScore:           10,
		BestLength:          14,
		BestDurationMillis:  12345,
		RunsPlayed:          3,
		TotalFoodEaten:      22,
		TotalPlayTimeMillis: 54321,
	}

	if err := repo.Save(context.Background(), in); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	out, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if out != in {
		t.Fatalf("round-trip mismatch: got=%+v want=%+v", out, in)
	}
}

func TestLoadMissingFileReturnsZeroProfile(t *testing.T) {
	dir := t.TempDir()
	repo := NewFileRepository(filepath.Join(dir, "missing.json"))

	out, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if out != (session.Profile{}) {
		t.Fatalf("expected zero profile for missing file: got=%+v", out)
	}
}
