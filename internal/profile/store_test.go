package profile

import (
	"path/filepath"
	"testing"

	"snake/internal/game"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	in := game.Profile{
		BestScore:           10,
		BestLength:          14,
		BestDurationMillis:  12345,
		RunsPlayed:          3,
		TotalFoodEaten:      22,
		TotalPlayTimeMillis: 54321,
	}

	if err := Save(path, in); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	out, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if out != in {
		t.Fatalf("round-trip mismatch: got=%+v want=%+v", out, in)
	}
}
