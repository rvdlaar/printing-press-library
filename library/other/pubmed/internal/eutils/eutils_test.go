// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.

package eutils

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestArticleAuthorsNeverNull(t *testing.T) {
	// A record with no authors (editorial/erratum/dataset) must still emit
	// "authors":[] so widget consumers can .map()/.length safely.
	a := articleFromDocSum("111", docSum{Title: "Editorial.", Source: "BMJ"})
	if a.Authors == nil {
		t.Fatal("Authors is nil; want empty non-nil slice")
	}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"authors":[]`) {
		t.Errorf("expected authors:[] in JSON, got %s", raw)
	}
}

func TestComposeQuery(t *testing.T) {
	cases := []struct {
		name    string
		base    string
		filters []string
		want    string
	}{
		{"base only", "atrial fibrillation", nil, "(atrial fibrillation)"},
		{"one filter", "sepsis", []string{"humans[MeSH Terms]"}, "(sepsis) AND (humans[MeSH Terms])"},
		{"skips empty", "sepsis", []string{"", "english[Language]", ""}, "(sepsis) AND (english[Language])"},
		{"two filters", "x", []string{"a", "b"}, "(x) AND (a) AND (b)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ComposeQuery(tc.base, tc.filters...); got != tc.want {
				t.Errorf("ComposeQuery(%q, %v) = %q, want %q", tc.base, tc.filters, got, tc.want)
			}
		})
	}
}

func TestClinicalFilter(t *testing.T) {
	cases := []struct {
		category, scope string
		wantContains    string
		wantErr         bool
	}{
		{"therapy", "narrow", "randomized controlled trial[Publication Type]", false},
		{"therapy", "broad", "clinical trial[Publication Type]", false},
		{"therapy", "", "clinical trial[Publication Type]", false}, // defaults to broad
		{"diagnosis", "narrow", "specificity[Title/Abstract]", false},
		{"prognosis", "broad", "prognos*[Text Word]", false},
		{"prediction", "narrow", "validation[Title/Abstract]", false},
		{"reviews", "", "systematic[sb]", false},
		{"reviews", "narrow", "systematic[sb]", false},                                // scope ignored
		{"THERAPY", "NARROW", "randomized controlled trial[Publication Type]", false}, // case-insensitive
		{"nonsense", "broad", "", true},
		{"therapy", "sideways", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.category+"/"+tc.scope, func(t *testing.T) {
			got, err := ClinicalFilter(tc.category, tc.scope)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q/%q, got none", tc.category, tc.scope)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(got, tc.wantContains) {
				t.Errorf("filter for %q/%q = %q, want substring %q", tc.category, tc.scope, got, tc.wantContains)
			}
		})
	}
}

func TestScopeFilters(t *testing.T) {
	got, err := ScopeFilters(ScopeOpts{Humans: true, English: true, Age: "aged", Sex: "female"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(got, " || ")
	for _, want := range []string{"humans[MeSH Terms]", "english[Language]", "aged[MeSH Terms]", "female[MeSH Terms]"} {
		if !strings.Contains(joined, want) {
			t.Errorf("scope filters %v missing %q", got, want)
		}
	}

	if _, err := ScopeFilters(ScopeOpts{Age: "geriatric"}); err == nil {
		t.Error("expected error for unknown age group")
	}
	if _, err := ScopeFilters(ScopeOpts{Sex: "other"}); err == nil {
		t.Error("expected error for unknown sex")
	}
	if _, err := ScopeFilters(ScopeOpts{Species: "plant"}); err == nil {
		t.Error("expected error for unknown species")
	}
	empty, err := ScopeFilters(ScopeOpts{})
	if err != nil || len(empty) != 0 {
		t.Errorf("empty scope should yield no filters, got %v err %v", empty, err)
	}
}

func TestClassifyEvidence(t *testing.T) {
	cases := []struct {
		pubTypes []string
		want     string
	}{
		{[]string{"Journal Article", "Meta-Analysis", "Review"}, "meta-analysis"}, // strongest wins
		{[]string{"Systematic Review"}, "systematic-review"},
		{[]string{"Randomized Controlled Trial", "Journal Article"}, "randomized-controlled-trial"},
		{[]string{"Clinical Trial"}, "clinical-trial"},
		{[]string{"Practice Guideline"}, "guideline"},
		{[]string{"Review"}, "review"},
		{[]string{"Case Reports"}, "case-report"},
		{[]string{"Journal Article"}, "article"},
		{nil, ""},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.pubTypes, "+"), func(t *testing.T) {
			if got := ClassifyEvidence(tc.pubTypes); got != tc.want {
				t.Errorf("ClassifyEvidence(%v) = %q, want %q", tc.pubTypes, got, tc.want)
			}
		})
	}
}

func TestParseYear(t *testing.T) {
	cases := []struct {
		sortPub, pubDate string
		want             int
	}{
		{"2026/07/01 00:00", "2026 Jul 1", 2026},
		{"", "2019 Dec", 2019},
		{"", "2015", 2015},
		{"", "n/a", 0},
	}
	for _, tc := range cases {
		if got := parseYear(tc.sortPub, tc.pubDate); got != tc.want {
			t.Errorf("parseYear(%q,%q) = %d, want %d", tc.sortPub, tc.pubDate, got, tc.want)
		}
	}
}

func TestArticleFromDocSum(t *testing.T) {
	ds := docSum{
		Title:   "A trial of something.",
		Source:  "N Engl J Med",
		PubDate: "2024 Jan 5",
		SortPub: "2024/01/05 00:00",
		PubType: []string{"Randomized Controlled Trial", "Journal Article"},
	}
	ds.Authors = append(ds.Authors, struct {
		Name string `json:"name"`
	}{Name: "Smith J"})
	ds.ArticleIDs = append(ds.ArticleIDs,
		struct {
			IDType string `json:"idtype"`
			Value  string `json:"value"`
		}{IDType: "doi", Value: "10.1056/x"},
		struct {
			IDType string `json:"idtype"`
			Value  string `json:"value"`
		}{IDType: "pmc", Value: "PMC12345"},
	)
	a := articleFromDocSum("40012345", ds)
	if a.PMID != "40012345" || a.DOI != "10.1056/x" || a.PMCID != "PMC12345" {
		t.Errorf("identity fields wrong: %+v", a)
	}
	if a.Year != 2024 {
		t.Errorf("year = %d, want 2024", a.Year)
	}
	if a.EvidenceLevel != "randomized-controlled-trial" {
		t.Errorf("evidence = %q", a.EvidenceLevel)
	}
	if a.URL != "https://pubmed.ncbi.nlm.nih.gov/40012345/" {
		t.Errorf("url = %q", a.URL)
	}
	if a.PMCURL != "https://www.ncbi.nlm.nih.gov/pmc/articles/PMC12345/" {
		t.Errorf("pmc url = %q", a.PMCURL)
	}
	if len(a.Authors) != 1 || a.Authors[0] != "Smith J" {
		t.Errorf("authors = %v", a.Authors)
	}
}

func TestParseAbstractXML(t *testing.T) {
	xmlBody := []byte(`<?xml version="1.0"?>
<PubmedArticleSet>
  <PubmedArticle>
    <MedlineCitation>
      <PMID Version="1">111</PMID>
      <Article>
        <Abstract>
          <AbstractText Label="BACKGROUND">First part.</AbstractText>
          <AbstractText Label="RESULTS">Second part.</AbstractText>
        </Abstract>
      </Article>
    </MedlineCitation>
  </PubmedArticle>
  <PubmedArticle>
    <MedlineCitation>
      <PMID Version="1">222</PMID>
      <Article>
        <Abstract>
          <AbstractText>Plain abstract.</AbstractText>
        </Abstract>
      </Article>
    </MedlineCitation>
  </PubmedArticle>
</PubmedArticleSet>`)
	got, err := parseAbstractXML(xmlBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got["111"], "BACKGROUND: First part.") || !strings.Contains(got["111"], "RESULTS: Second part.") {
		t.Errorf("labeled abstract wrong: %q", got["111"])
	}
	if got["222"] != "Plain abstract." {
		t.Errorf("plain abstract = %q", got["222"])
	}
}
