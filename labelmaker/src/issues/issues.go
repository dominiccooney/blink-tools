// Package issues parses issues from the Chromium issue tracker.
//
// Most issues can be downloaded from this URL:
// https://code.google.com/feeds/issues/p/chromium/issues/full?
//
// The set of issues returned can be refined with query string
// parameters; here are some useful ones:
//
// * (crbug.com query parameters, eg q=-is:open to skip open issues)
// * updated-min=YYYY-mm-ddT00:00:00
// * updated-max=...
// * alt=json (for JSON instead of XML)
// * start-index=26 (first issue on page)
// * max-results=25 (page size)
//
// See https://code.google.com/p/support/wiki/IssueTrackerAPI and
// https://code.google.com/p/chromium/issues/searchtips for more.
package issues

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
)

// State indicates whether an issue is open or closed.
type State bool

const (
	StateClosed State = false
	StateOpen         = true
)

type Status string

const (
	StatusUnconfirmed        Status = "Unconfirmed"
	StatusUntriaged                 = "Untriaged"
	StatusAvailable                 = "Available"
	StatusAssigned                  = "Assigned"
	StatusStarted                   = "Started"
	StatusExternalDependency        = "ExternalDependency"
	StatusFixed                     = "Fixed"
	StatusVerified                  = "Verified"
	StatusDuplicate                 = "Duplicate"
	StatusWontFix                   = "WontFix"
	StatusArchived                  = "Archived"
)

// IssueLabels is a set of string labels. Labels are an open
// taxonomy. Some tools attach special meaning to specific
// labels. Some labels have structure, for example, Cr-X-Y refers to
// the X component's Y subcomponent.
type Labels []string

type Issue struct {
	Id          int
	Title       string
	Content     string
	State       State
	Status      Status
	IssueLabels Labels
}

func parseIssueDecodedJson(entry map[string]interface{}) (*Issue, error) {
	p := newIssueParser(entry)
	issue := &Issue{
		p.id(),
		p.title(),
		p.content(),
		p.state(),
		p.status(),
		p.labels(),
	}
	if p.err != nil {
		return nil, p.err
	}
	return issue, nil
}

type issueParser struct {
	entry map[string]interface{}
	err   error
}

func newIssueParser(entry map[string]interface{}) *issueParser {
	return &issueParser{entry, nil}
}

var issueParserRegexp = regexp.MustCompile(`^http://code\.google\.com/feeds/issues/p/chromium/issues/full/(\d+)$`)

func (p *issueParser) id() int {
	s := p.entry["id"].(map[string]interface{})["$t"].(string)
	if !issueParserRegexp.MatchString(s) {
		p.err = fmt.Errorf("Could not match \"%s\" as an issue ID", s)
		return -1
	}
	var id int
	id, p.err = strconv.Atoi(issueParserRegexp.ReplaceAllString(s, "$1"))
	return id
}

func (p *issueParser) title() string {
	return p.entry["title"].(map[string]interface{})["$t"].(string)
}

func (p *issueParser) content() string {
	return html.UnescapeString(p.entry["content"].(map[string]interface{})["$t"].(string))
}

func (p *issueParser) state() State {
	s := p.entry["issues$state"].(map[string]interface{})["$t"].(string)
	switch s {
	case "closed":
		return StateClosed
	case "open":
		return StateOpen
	default:
		p.err = fmt.Errorf("Unrecognized state \"%s\"", s)
		return StateClosed
	}
}

func (p *issueParser) status() Status {
	s := p.entry["issues$status"]
	if s == nil {
		// Issue 475886 has no status, but viewed through the
		// FE it is apparently untriaged:
		// https://code.google.com/feeds/issues/p/chromium/issues/full/?q=475886&alt=json
		// https://crbug.com/475886
		return StatusUntriaged
	}
	return Status(s.(map[string]interface{})["$t"].(string))
}

func (p *issueParser) labels() Labels {
	var ls []string = nil
	labelsJson := p.entry["issues$label"]
	if labelsJson == nil {
		return nil
	}
	for _, value := range labelsJson.([]interface{}) {
		ls = append(ls, value.(map[string]interface{})["$t"].(string))
	}
	return ls
}

func ParseIssuesJson(content []byte) ([]*Issue, error) {
	var doc interface{}
	err := json.Unmarshal(content, &doc)
	if err != nil {
		return nil, err
	}
	return parseIssuesDecodedJson(doc)
}

func feed(doc map[string]interface{}) map[string]interface{} {
	return doc["feed"].(map[string]interface{})
}

func entries(feed map[string]interface{}) []interface{} {
	return feed["entry"].([]interface{})
}

func parseIssuesDecodedJson(doc interface{}) ([]*Issue, error) {
	var issues []*Issue
	for _, entry := range entries(feed(doc.(map[string]interface{}))) {
		issue, err := parseIssueDecodedJson(entry.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}
