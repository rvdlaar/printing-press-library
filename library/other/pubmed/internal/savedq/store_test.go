// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.

package savedq

import (
	"testing"
)

// withStore points the store at a temp dir via HOME so tests don't touch the
// real data directory. cliutil.DataDir honors platform config; setting HOME (and
// XDG on Linux) covers the common resolution paths, and Open falls back to
// $HOME/.local/share regardless.
func withStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)
	s, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return s
}

func TestAddGetRemove(t *testing.T) {
	s := withStore(t)
	if _, err := s.Add("AF-Anticoag", "atrial fibrillation AND anticoagulation"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// Name is normalized (lowercased) for keying.
	got, ok := s.Get("af-anticoag")
	if !ok || got.Term != "atrial fibrillation AND anticoagulation" {
		t.Fatalf("Get after Add: %+v ok=%v", got, ok)
	}
	// Reload from disk to confirm persistence.
	s2, err := Open()
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if _, ok := s2.Get("af-anticoag"); !ok {
		t.Fatal("query did not persist across reopen")
	}
	removed, err := s2.Remove("af-anticoag")
	if err != nil || !removed {
		t.Fatalf("Remove: removed=%v err=%v", removed, err)
	}
	if _, ok := s2.Get("af-anticoag"); ok {
		t.Fatal("query still present after Remove")
	}
}

func TestAddResetsBaselineOnTermChange(t *testing.T) {
	s := withStore(t)
	if _, err := s.Add("q", "term one"); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordRun("q", "term one", []string{"1", "2"}, 2); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get("q"); len(got.SeenPMIDs) != 2 {
		t.Fatalf("expected 2 seen, got %d", len(got.SeenPMIDs))
	}
	// Changing the term resets the watch baseline.
	if _, err := s.Add("q", "term two"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get("q"); len(got.SeenPMIDs) != 0 {
		t.Fatalf("term change should reset baseline, got %d seen", len(got.SeenPMIDs))
	}
}

func TestRecordRunUnionsAndPersistsTerm(t *testing.T) {
	s := withStore(t)
	// Ad-hoc watch (never added): RecordRun must persist the term so a re-run
	// survives (Finding 3 regression guard).
	if err := s.RecordRun("sepsis biomarkers", "sepsis biomarkers", []string{"1", "2"}, 10); err != nil {
		t.Fatal(err)
	}
	got, ok := s.Get("sepsis biomarkers")
	if !ok || got.Term != "sepsis biomarkers" {
		t.Fatalf("term not persisted for ad-hoc watch: %+v", got)
	}
	// Second run adds a new PMID; seen set is the union, no duplicates.
	if err := s.RecordRun("sepsis biomarkers", "sepsis biomarkers", []string{"2", "3"}, 11); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Get("sepsis biomarkers")
	seen := map[string]int{}
	for _, p := range got.SeenPMIDs {
		seen[p]++
	}
	for _, p := range []string{"1", "2", "3"} {
		if seen[p] != 1 {
			t.Errorf("PMID %s seen %d times, want exactly 1 (set: %v)", p, seen[p], got.SeenPMIDs)
		}
	}
}

func TestListEmptyIsNonNil(t *testing.T) {
	s := withStore(t)
	if got := s.List(); got == nil || len(got) != 0 {
		t.Fatalf("empty List should be non-nil zero-length, got %#v", got)
	}
}
