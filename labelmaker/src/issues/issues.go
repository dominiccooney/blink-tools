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
	"strconv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"html"
	"regexp"
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

// Labels is a set of string labels. Labels are an open taxonomy. Some
// tools attach special meaning to specific labels. Some labels have
// structure, for example, Cr-X-Y refers to the X component's Y
// subcomponent.
type Labels []string

type Issue struct {
	Id      int
	Title   string
	Content string
	State   State
	Status  Status
	Labels  Labels
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
	err error
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
	return Status(p.entry["issues$status"].(map[string]interface{})["$t"].(string))
}

func (p *issueParser) labels() Labels {
	var ls []string = nil
	for _, value := range p.entry["issues$label"].([]interface{}) {
		ls = append(ls, value.(map[string]interface{})["$t"].(string))
	}
	return ls
}

func parseIssuesJson(content []byte) ([]*Issue, error) {
        var doc interface{}
	err := json.Unmarshal(content, &doc)
	if err != nil {
		return nil, err
	}
	return parseIssuesDecodedJson(doc)
}

func parseIssuesDecodedJson(doc interface{}) ([]*Issue, error) {
	var issues []*Issue
	for _, entry := range doc.(map[string]interface{})["feed"].(map[string]interface{})["entry"].([]interface{}) {
		issue, err := parseIssueDecodedJson(entry.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

func ParseIssues() ([]*Issue, error) {
	// TODO: This just pulls in canned test data from one file.
	const filePath = "sample-issues.json"
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return parseIssuesJson(bytes)
}
