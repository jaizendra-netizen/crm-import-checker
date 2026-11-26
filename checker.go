// Package main checks a CSV before it is imported into a CRM: stable match keys, duplicates, empty key fields and stage names.
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/mail"
	"strings"
)

// DefaultStages are the lifecycle stages named on the RevOps service page: MQL, SQL and opportunity.
var DefaultStages = []string{"mql", "sql", "opportunity"}

type Issue struct {
	Row     int
	Kind    string
	Message string
}

type Options struct {
	Required []string // fields sales relies on; must not be empty
	Stages   []string // allowed values in the stage column (lower case)
}

func normDomain(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	return s
}

// Check reads CSV rows with a header and returns issues. Records are matched on domain and email, never on name.
func Check(r io.Reader, o Options) ([]Issue, int, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	recs, err := cr.ReadAll()
	if err != nil {
		return nil, 0, err
	}
	if len(recs) == 0 {
		return nil, 0, fmt.Errorf("empty file")
	}
	col := map[string]int{}
	for i, h := range recs[0] {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	get := func(rec []string, name string) string {
		if i, ok := col[name]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}
	var issues []Issue
	seenDomain := map[string]int{}
	seenEmail := map[string]int{}
	_, hasDomain := col["domain"]
	_, hasEmail := col["email"]
	if !hasDomain && !hasEmail {
		return nil, 0, fmt.Errorf("need a domain column or an email column to match records")
	}
	stages := o.Stages
	if len(stages) == 0 {
		stages = DefaultStages
	}
	for n, rec := range recs[1:] {
		row := n + 2
		if hasDomain {
			d := normDomain(get(rec, "domain"))
			if d == "" {
				issues = append(issues, Issue{row, "missing_key", "domain is empty"})
			} else if first, dup := seenDomain[d]; dup && !hasEmail {
				issues = append(issues, Issue{row, "duplicate", fmt.Sprintf("domain %s already on row %d", d, first)})
			} else if !dup {
				seenDomain[d] = row
			}
		}
		if hasEmail {
			e := strings.ToLower(get(rec, "email"))
			if e == "" {
				issues = append(issues, Issue{row, "missing_key", "email is empty"})
			} else if _, err := mail.ParseAddress(e); err != nil {
				issues = append(issues, Issue{row, "bad_email", e + " is not a valid address"})
			} else if first, dup := seenEmail[e]; dup {
				issues = append(issues, Issue{row, "duplicate", fmt.Sprintf("email %s already on row %d", e, first)})
			} else {
				seenEmail[e] = row
			}
		}
		for _, f := range o.Required {
			if get(rec, strings.ToLower(f)) == "" {
				issues = append(issues, Issue{row, "empty_field", f + " is empty"})
			}
		}
		if _, ok := col["stage"]; ok {
			s := strings.ToLower(get(rec, "stage"))
			if s != "" && !contains(stages, s) {
				issues = append(issues, Issue{row, "unknown_stage", fmt.Sprintf("stage %q is not one of %s", s, strings.Join(stages, ", "))})
			}
		}
	}
	return issues, len(recs) - 1, nil
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}
