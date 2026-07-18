// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored local store for named saved queries and per-query seen-PMID
// sets (powers `saved` and `watch`). Its own package so generate --force keeps it.

// Package savedq persists named PubMed queries and the PMIDs already seen for
// each, in a single JSON file under the CLI data directory.
package savedq

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/cliutil"
)

// Saved is one named query plus its watch state.
type Saved struct {
	Name      string    `json:"name"`
	Term      string    `json:"term"`
	LastRun   time.Time `json:"last_run,omitempty"`
	LastCount int       `json:"last_count,omitempty"`
	SeenPMIDs []string  `json:"seen_pmids,omitempty"`
}

// Store is the on-disk collection of saved queries.
type Store struct {
	path    string
	Queries map[string]Saved `json:"queries"`
}

func storePath() (string, error) {
	dir, err := cliutil.DataDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", fmt.Errorf("resolving data dir: %w", err)
		}
		dir = filepath.Join(home, ".local", "share", "pubmed-pp-cli")
	}
	return filepath.Join(dir, "saved-queries.json"), nil
}

// Open loads the store, creating an empty one if the file does not yet exist.
func Open() (*Store, error) {
	p, err := storePath()
	if err != nil {
		return nil, err
	}
	s := &Store{path: p, Queries: map[string]Saved{}}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("reading saved-queries store: %w", err)
	}
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing saved-queries store: %w", err)
	}
	if s.Queries == nil {
		s.Queries = map[string]Saved{}
	}
	return s, nil
}

// save atomically writes the store to disk via a unique temp file + rename.
// A unique temp name (os.CreateTemp) keeps two concurrent writers from sharing
// one ".tmp" path and clobbering each other's rename. This does not close the
// read-modify-write lost-update window inherent to a whole-file store — callers
// are expected to invoke store operations serially (single-writer discipline).
func (s *Store) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(s.path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp store file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing saved-queries store: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp store file: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("finalizing saved-queries store: %w", err)
	}
	return nil
}

// NormalizeName lowercases and trims a query name for stable keying.
func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// List returns all saved queries sorted by name.
func (s *Store) List() []Saved {
	out := make([]Saved, 0, len(s.Queries))
	for _, q := range s.Queries {
		out = append(out, q)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get returns a saved query by name.
func (s *Store) Get(name string) (Saved, bool) {
	q, ok := s.Queries[NormalizeName(name)]
	return q, ok
}

// Add creates or replaces a saved query's term (preserving any existing watch
// state when the term is unchanged).
func (s *Store) Add(name, term string) (Saved, error) {
	key := NormalizeName(name)
	if key == "" {
		return Saved{}, fmt.Errorf("query name is required")
	}
	if strings.TrimSpace(term) == "" {
		return Saved{}, fmt.Errorf("query term is required")
	}
	q := s.Queries[key]
	if q.Term != term {
		q.SeenPMIDs = nil // term changed → reset watch baseline
		q.LastRun = time.Time{}
		q.LastCount = 0
	}
	q.Name = key
	q.Term = term
	s.Queries[key] = q
	if err := s.save(); err != nil {
		return Saved{}, err
	}
	return q, nil
}

// Remove deletes a saved query. Returns false if it did not exist.
func (s *Store) Remove(name string) (bool, error) {
	key := NormalizeName(name)
	if _, ok := s.Queries[key]; !ok {
		return false, nil
	}
	delete(s.Queries, key)
	return true, s.save()
}

// RecordRun updates a saved query's watch state after a run: marks the given
// PMIDs as seen, persists the term (so ad-hoc watches survive a re-run), and
// stamps last-run/last-count.
func (s *Store) RecordRun(name, term string, allPMIDs []string, count int) error {
	key := NormalizeName(name)
	q := s.Queries[key]
	q.Name = key
	if term != "" {
		q.Term = term
	}
	seen := map[string]bool{}
	for _, p := range q.SeenPMIDs {
		seen[p] = true
	}
	for _, p := range allPMIDs {
		if !seen[p] {
			seen[p] = true
			q.SeenPMIDs = append(q.SeenPMIDs, p)
		}
	}
	q.LastRun = time.Now().UTC()
	q.LastCount = count
	s.Queries[key] = q
	return s.save()
}

// Path returns the on-disk location of the store (for doctor/diagnostics).
func (s *Store) Path() string { return s.path }
