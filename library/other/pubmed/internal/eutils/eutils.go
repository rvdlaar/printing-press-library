// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored shared logic for PubMed E-utilities novel commands.
// Lives in its own package so `generate --force` preserves it as a whole unit.

// Package eutils wraps NCBI E-utilities (esearch/esummary/efetch) with clean,
// widget-ready types plus the clinical-query, scope, and open-access filter
// logic that the novel pubmed-pp-cli commands build on.
package eutils

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Client is the minimal surface the generated internal/client.Client satisfies.
// Declaring it here keeps eutils decoupled from the CLI package and testable
// with a fake.
type Client interface {
	Get(ctx context.Context, path string, params map[string]string) (json.RawMessage, error)
}

// Article is a clean, widget-ready PubMed record. Fields the API omits stay at
// their zero value and are dropped from JSON via omitempty where appropriate.
type Article struct {
	PMID          string   `json:"pmid"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Journal       string   `json:"journal"`
	PubDate       string   `json:"pub_date"`
	Year          int      `json:"year,omitempty"`
	DOI           string   `json:"doi,omitempty"`
	PMCID         string   `json:"pmcid,omitempty"`
	EvidenceLevel string   `json:"evidence_level,omitempty"`
	PubTypes      []string `json:"pub_types,omitempty"`
	URL           string   `json:"url"`
	PMCURL        string   `json:"pmc_url,omitempty"`
	Abstract      string   `json:"abstract,omitempty"`
}

// SearchResult is the outcome of an esearch call.
type SearchResult struct {
	Term             string   `json:"term"`
	Count            int      `json:"count"`
	PMIDs            []string `json:"pmids"`
	QueryTranslation string   `json:"query_translation,omitempty"`
}

// SearchOpts carries optional esearch parameters.
type SearchOpts struct {
	Retmax   int
	Retstart int
	Sort     string
	Datetype string
	Mindate  string
	Maxdate  string
	Reldate  int
}

const (
	pubmedArticleURL = "https://pubmed.ncbi.nlm.nih.gov/"
	pmcArticleURL    = "https://www.ncbi.nlm.nih.gov/pmc/articles/"
	toolName         = "pubmed-pp-cli"
)

// withBase adds the NCBI-recommended tool/email identifiers and the optional
// API key (NCBI_API_KEY raises the rate limit from 3 to 10 req/s). Reading the
// env here means every E-utilities call picks up the key without each command
// having to thread it through.
func withBase(p map[string]string) map[string]string {
	if p == nil {
		p = map[string]string{}
	}
	p["tool"] = toolName
	if k := strings.TrimSpace(os.Getenv("NCBI_API_KEY")); k != "" {
		p["api_key"] = k
	}
	if e := strings.TrimSpace(os.Getenv("NCBI_EMAIL")); e != "" {
		p["email"] = e
	}
	return p
}

// HasAPIKey reports whether an NCBI API key is configured (used to pick a
// throttle interval for fan-out commands).
func HasAPIKey() bool { return strings.TrimSpace(os.Getenv("NCBI_API_KEY")) != "" }

// ThrottleInterval returns a safe per-request spacing for fan-out commands:
// ~10 req/s with an API key, ~3 req/s without.
func ThrottleInterval() time.Duration {
	if HasAPIKey() {
		return 110 * time.Millisecond
	}
	return 350 * time.Millisecond
}

// Search runs esearch and returns the PMID list, total count, and PubMed's
// query translation.
func Search(ctx context.Context, c Client, term string, opts SearchOpts) (SearchResult, error) {
	if strings.TrimSpace(term) == "" {
		return SearchResult{}, fmt.Errorf("empty search term")
	}
	retmax := opts.Retmax
	if retmax <= 0 {
		retmax = 20
	}
	params := withBase(map[string]string{
		"db":      "pubmed",
		"term":    term,
		"retmode": "json",
		"retmax":  strconv.Itoa(retmax),
	})
	if opts.Retstart > 0 {
		params["retstart"] = strconv.Itoa(opts.Retstart)
	}
	if opts.Sort != "" {
		params["sort"] = opts.Sort
	}
	if opts.Datetype != "" {
		params["datetype"] = opts.Datetype
	}
	if opts.Mindate != "" {
		params["mindate"] = opts.Mindate
	}
	if opts.Maxdate != "" {
		params["maxdate"] = opts.Maxdate
	}
	if opts.Reldate > 0 {
		params["reldate"] = strconv.Itoa(opts.Reldate)
	}
	data, err := c.Get(ctx, "/esearch.fcgi", params)
	if err != nil {
		return SearchResult{}, err
	}
	var env struct {
		ESearchResult struct {
			Count            string   `json:"count"`
			IDList           []string `json:"idlist"`
			QueryTranslation string   `json:"querytranslation"`
			Error            string   `json:"error"`
		} `json:"esearchresult"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return SearchResult{}, fmt.Errorf("parsing esearch response: %w", err)
	}
	if env.ESearchResult.Error != "" {
		return SearchResult{}, fmt.Errorf("esearch error: %s", env.ESearchResult.Error)
	}
	count, _ := strconv.Atoi(env.ESearchResult.Count)
	return SearchResult{
		Term:             term,
		Count:            count,
		PMIDs:            env.ESearchResult.IDList,
		QueryTranslation: env.ESearchResult.QueryTranslation,
	}, nil
}

// SearchCount runs esearch with retmax=0 and returns only the total count.
// Used by fan-out commands (trend, cooccur) that need counts, not records.
func SearchCount(ctx context.Context, c Client, term string) (int, error) {
	params := withBase(map[string]string{
		"db":      "pubmed",
		"term":    term,
		"retmode": "json",
		"retmax":  "0",
	})
	data, err := c.Get(ctx, "/esearch.fcgi", params)
	if err != nil {
		return 0, err
	}
	var env struct {
		ESearchResult struct {
			Count string `json:"count"`
			Error string `json:"error"`
		} `json:"esearchresult"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return 0, fmt.Errorf("parsing esearch response: %w", err)
	}
	if env.ESearchResult.Error != "" {
		return 0, fmt.Errorf("esearch error: %s", env.ESearchResult.Error)
	}
	count, _ := strconv.Atoi(env.ESearchResult.Count)
	return count, nil
}

type docSum struct {
	UID     string `json:"uid"`
	Title   string `json:"title"`
	Source  string `json:"source"`
	PubDate string `json:"pubdate"`
	SortPub string `json:"sortpubdate"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
	ArticleIDs []struct {
		IDType string `json:"idtype"`
		Value  string `json:"value"`
	} `json:"articleids"`
	PubType []string `json:"pubtype"`
}

// Summaries runs esummary for the given PMIDs and returns clean Article cards,
// preserving the input PMID order.
func Summaries(ctx context.Context, c Client, pmids []string) ([]Article, error) {
	if len(pmids) == 0 {
		return []Article{}, nil
	}
	params := withBase(map[string]string{
		"db":      "pubmed",
		"id":      strings.Join(pmids, ","),
		"retmode": "json",
	})
	data, err := c.Get(ctx, "/esummary.fcgi", params)
	if err != nil {
		return nil, err
	}
	var env struct {
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parsing esummary response: %w", err)
	}
	articles := make([]Article, 0, len(pmids))
	for _, pmid := range pmids {
		raw, ok := env.Result[pmid]
		if !ok {
			continue
		}
		var ds docSum
		if err := json.Unmarshal(raw, &ds); err != nil {
			continue
		}
		articles = append(articles, articleFromDocSum(pmid, ds))
	}
	return articles, nil
}

func articleFromDocSum(pmid string, ds docSum) Article {
	a := Article{
		PMID:     pmid,
		Title:    strings.TrimSpace(ds.Title),
		Journal:  ds.Source,
		PubDate:  ds.PubDate,
		PubTypes: ds.PubType,
		URL:      pubmedArticleURL + pmid + "/",
	}
	// Always emit a JSON array for authors, never null, so widget consumers can
	// safely .map()/.length over it even for author-less records (editorials,
	// errata, dataset entries).
	a.Authors = make([]string, 0, len(ds.Authors))
	for _, au := range ds.Authors {
		if au.Name != "" {
			a.Authors = append(a.Authors, au.Name)
		}
	}
	for _, id := range ds.ArticleIDs {
		switch strings.ToLower(id.IDType) {
		case "doi":
			a.DOI = id.Value
		case "pmc":
			a.PMCID = id.Value
		}
	}
	if a.PMCID != "" {
		a.PMCURL = pmcArticleURL + a.PMCID + "/"
	}
	a.Year = parseYear(ds.SortPub, ds.PubDate)
	a.EvidenceLevel = ClassifyEvidence(ds.PubType)
	return a
}

// parseYear extracts a 4-digit year from sortpubdate ("YYYY/MM/DD ...") or
// pubdate ("2026 Jul 1" / "2026").
func parseYear(sortPub, pubDate string) int {
	for _, s := range []string{sortPub, pubDate} {
		s = strings.TrimSpace(s)
		if len(s) >= 4 {
			if y, err := strconv.Atoi(s[:4]); err == nil && y > 1500 && y < 3000 {
				return y
			}
		}
	}
	return 0
}

// Abstracts runs efetch (rettype=abstract, retmode=xml) and returns a
// PMID→abstract-text map. efetch has no JSON mode, so the XML is parsed here.
func Abstracts(ctx context.Context, c Client, pmids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(pmids) == 0 {
		return out, nil
	}
	params := withBase(map[string]string{
		"db":      "pubmed",
		"id":      strings.Join(pmids, ","),
		"rettype": "abstract",
		"retmode": "xml",
	})
	data, err := c.Get(ctx, "/efetch.fcgi", params)
	if err != nil {
		return nil, err
	}
	return parseAbstractXML(data)
}

type pubmedArticleSet struct {
	Articles []struct {
		PMID     string `xml:"MedlineCitation>PMID"`
		Abstract []struct {
			Label string `xml:"Label,attr"`
			Text  string `xml:",chardata"`
		} `xml:"MedlineCitation>Article>Abstract>AbstractText"`
	} `xml:"PubmedArticle"`
}

func parseAbstractXML(data []byte) (map[string]string, error) {
	out := map[string]string{}
	var set pubmedArticleSet
	if err := xml.Unmarshal(data, &set); err != nil {
		return nil, fmt.Errorf("parsing efetch XML: %w", err)
	}
	for _, art := range set.Articles {
		pmid := strings.TrimSpace(art.PMID)
		if pmid == "" {
			continue
		}
		var parts []string
		for _, ab := range art.Abstract {
			seg := strings.TrimSpace(ab.Text)
			if seg == "" {
				continue
			}
			if label := strings.TrimSpace(ab.Label); label != "" {
				parts = append(parts, label+": "+seg)
			} else {
				parts = append(parts, seg)
			}
		}
		out[pmid] = strings.Join(parts, "\n\n")
	}
	return out, nil
}

// AttachAbstracts fills the Abstract field on each article in place.
func AttachAbstracts(ctx context.Context, c Client, articles []Article) error {
	if len(articles) == 0 {
		return nil
	}
	pmids := make([]string, 0, len(articles))
	for _, a := range articles {
		pmids = append(pmids, a.PMID)
	}
	abs, err := Abstracts(ctx, c, pmids)
	if err != nil {
		return err
	}
	for i := range articles {
		if txt, ok := abs[articles[i].PMID]; ok {
			articles[i].Abstract = txt
		}
	}
	return nil
}

// ---- Query composition & filters ----

// ComposeQuery ANDs a base term with any non-empty filter clauses, wrapping each
// in parentheses so precedence is unambiguous.
func ComposeQuery(base string, filters ...string) string {
	q := "(" + strings.TrimSpace(base) + ")"
	for _, f := range filters {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		q += " AND (" + f + ")"
	}
	return q
}

// clinicalCategories are the canonical PubMed Clinical Queries (Haynes) filter
// strategies, keyed by category then scope (broad=sensitive, narrow=specific).
var clinicalCategories = map[string]map[string]string{
	"therapy": {
		"broad":  "clinical[Title/Abstract] AND trial[Title/Abstract] OR clinical trials as topic[MeSH Terms] OR clinical trial[Publication Type] OR random*[Title/Abstract] OR random allocation[MeSH Terms] OR therapeutic use[MeSH Subheading]",
		"narrow": "randomized controlled trial[Publication Type] OR (randomized[Title/Abstract] AND controlled[Title/Abstract] AND trial[Title/Abstract])",
	},
	"diagnosis": {
		"broad":  "sensitiv*[Title/Abstract] OR sensitivity and specificity[MeSH Terms] OR (predictive[Title/Abstract] AND value*[Title/Abstract]) OR accuracy[Title/Abstract]",
		"narrow": "specificity[Title/Abstract]",
	},
	"etiology": {
		"broad":  "risk*[Title/Abstract] OR risk*[MeSH:noexp] OR cohort studies[MeSH Terms] OR group[Text Word] OR (odds[Text Word] AND ratio*[Text Word]) OR (relative[Text Word] AND risk*[Text Word]) OR (case*[Text Word] AND control*[Text Word])",
		"narrow": "(relative[Title/Abstract] AND risk*[Title/Abstract]) OR (relative risk[Text Word]) OR risks[Text Word] OR cohort studies[MeSH:noexp] OR (case*[Title/Abstract] AND control*[Title/Abstract]) OR case-control studies[MeSH:noexp]",
	},
	"prognosis": {
		"broad":  "incidence[MeSH:noexp] OR mortality[MeSH Terms] OR follow up studies[MeSH:noexp] OR prognos*[Text Word] OR predict*[Text Word] OR course*[Text Word]",
		"narrow": "prognos*[Title/Abstract] OR (first[Title/Abstract] AND episode[Title/Abstract]) OR cohort[Title/Abstract]",
	},
	"prediction": {
		"broad":  "predict*[Title/Abstract] OR scor*[Title/Abstract] OR observ*[Title/Abstract] OR observer variation[MeSH Terms]",
		"narrow": "validation[Title/Abstract] OR validate[Title/Abstract]",
	},
}

// ClinicalCategories returns the supported clinical-query categories, sorted.
func ClinicalCategories() []string {
	out := make([]string, 0, len(clinicalCategories)+1)
	for k := range clinicalCategories {
		out = append(out, k)
	}
	out = append(out, "reviews")
	sort.Strings(out)
	return out
}

// ClinicalFilter returns the filter clause for a Clinical Queries category and
// scope. Category "reviews" returns the systematic-review subset and ignores
// scope. Scope defaults to "broad" when empty.
func ClinicalFilter(category, scope string) (string, error) {
	category = strings.ToLower(strings.TrimSpace(category))
	if category == "reviews" || category == "systematic" {
		return "systematic[sb]", nil
	}
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		scope = "broad"
	}
	if scope == "sensitive" {
		scope = "broad"
	}
	if scope == "specific" {
		scope = "narrow"
	}
	m, ok := clinicalCategories[category]
	if !ok {
		return "", fmt.Errorf("unknown category %q (valid: %s)", category, strings.Join(ClinicalCategories(), ", "))
	}
	clause, ok := m[scope]
	if !ok {
		return "", fmt.Errorf("unknown scope %q (valid: broad, narrow)", scope)
	}
	return clause, nil
}

// OpenAccessFilter returns the free-full-text subset filter.
func OpenAccessFilter() string { return "free full text[sb]" }

var ageGroupMeSH = map[string]string{
	"newborn":     "infant, newborn[MeSH Terms]",
	"infant":      "infant[MeSH Terms]",
	"child":       "child[MeSH Terms]",
	"adolescent":  "adolescent[MeSH Terms]",
	"adult":       "adult[MeSH Terms]",
	"middle-aged": "middle aged[MeSH Terms]",
	"aged":        "aged[MeSH Terms]",
}

// AgeGroups returns the supported --age values, sorted.
func AgeGroups() []string {
	out := make([]string, 0, len(ageGroupMeSH))
	for k := range ageGroupMeSH {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ScopeOpts carries the clinical scope-filter flags.
type ScopeOpts struct {
	Humans  bool
	English bool
	Age     string
	Sex     string
	Species string
}

// ScopeFilters converts scope options into PubMed field-tag clauses. Returns an
// error for an unrecognized age/sex/species so a typo fails loudly instead of
// silently returning unfiltered results.
func ScopeFilters(o ScopeOpts) ([]string, error) {
	var f []string
	if o.Humans {
		f = append(f, "humans[MeSH Terms]")
	}
	if sp := strings.ToLower(strings.TrimSpace(o.Species)); sp != "" {
		switch sp {
		case "humans", "human":
			f = append(f, "humans[MeSH Terms]")
		case "animals", "animal":
			f = append(f, "animals[MeSH Terms]")
		default:
			return nil, fmt.Errorf("unknown --species %q (valid: humans, animals)", o.Species)
		}
	}
	if o.English {
		f = append(f, "english[Language]")
	}
	if age := strings.ToLower(strings.TrimSpace(o.Age)); age != "" {
		clause, ok := ageGroupMeSH[age]
		if !ok {
			return nil, fmt.Errorf("unknown --age %q (valid: %s)", o.Age, strings.Join(AgeGroups(), ", "))
		}
		f = append(f, clause)
	}
	if sex := strings.ToLower(strings.TrimSpace(o.Sex)); sex != "" {
		switch sex {
		case "male", "m":
			f = append(f, "male[MeSH Terms]")
		case "female", "f":
			f = append(f, "female[MeSH Terms]")
		default:
			return nil, fmt.Errorf("unknown --sex %q (valid: male, female)", o.Sex)
		}
	}
	return f, nil
}

// evidenceRank orders evidence levels from strongest to weakest so ClassifyEvidence
// returns the highest-strength matching type.
var evidenceRank = []struct {
	needle string
	level  string
}{
	{"meta-analysis", "meta-analysis"},
	{"systematic review", "systematic-review"},
	{"practice guideline", "guideline"},
	{"guideline", "guideline"},
	{"randomized controlled trial", "randomized-controlled-trial"},
	{"clinical trial", "clinical-trial"},
	{"review", "review"},
	{"case reports", "case-report"},
	{"comparative study", "comparative-study"},
	{"observational study", "observational-study"},
	{"editorial", "editorial"},
	{"letter", "letter"},
}

// ClassifyEvidence maps a record's publication types to a single evidence-level
// badge, choosing the strongest matching type.
func ClassifyEvidence(pubTypes []string) string {
	if len(pubTypes) == 0 {
		return ""
	}
	joined := strings.ToLower(strings.Join(pubTypes, "|"))
	for _, r := range evidenceRank {
		if strings.Contains(joined, r.needle) {
			return r.level
		}
	}
	return "article"
}

// EvidenceLevels lists the badges ClassifyEvidence can emit, for --evidence
// validation and help text.
func EvidenceLevels() []string {
	return []string{
		"meta-analysis", "systematic-review", "guideline",
		"randomized-controlled-trial", "clinical-trial", "review",
		"case-report", "comparative-study", "observational-study",
	}
}
