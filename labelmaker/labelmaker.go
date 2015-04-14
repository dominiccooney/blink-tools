// Machine larnin' on the Chromium issue tracker.

package main

import (
	"strconv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"html"
	"regexp"
)

type state bool

const (
	stateClosed state = false
	stateOpen         = true
)

type status string

const (
	statusUnconfirmed        status = "Unconfirmed"
	statusUntriaged                 = "Untriaged"
	statusAvailable                 = "Available"
	statusAssigned                  = "Assigned"
	statusStarted                   = "Started"
	statusExternalDependency        = "ExternalDependency"
	statusFixed                     = "Fixed"
	statusVerified                  = "Verified"
	statusDuplicate                 = "Duplicate"
	statusWontFix                   = "WontFix"
	statusArchived                  = "Archived"
)

type labels []string

type issue struct {
	id      int
	title   string
	content string
	state   state
	status  status
	labels  labels
}

func parseIssueDecodedJson(entry map[string]interface{}) (*issue, error) {
	p := newIssueParser(entry)
	issue := &issue{
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

func (p *issueParser) state() state {
	s := p.entry["issues$state"].(map[string]interface{})["$t"].(string)
	switch s {
	case "closed":
		return stateClosed
	case "open":
		return stateOpen
	default:
		p.err = fmt.Errorf("Unrecognized state \"%s\"", s)
		return stateClosed
	}
}

func (p *issueParser) status() status {
	return status(p.entry["issues$status"].(map[string]interface{})["$t"].(string))
}

func (p *issueParser) labels() []string {
	var ls []string = nil
	for _, value := range p.entry["issues$label"].([]interface{}) {
		ls = append(ls, value.(map[string]interface{})["$t"].(string))
	}
	return ls
}

func parseIssuesJson(content []byte) ([]*issue, error) {
        var doc interface{}
	err := json.Unmarshal(content, &doc)
	if err != nil {
		return nil, err
	}
	return parseIssuesDecodedJson(doc)
}

func parseIssuesDecodedJson(doc interface{}) ([]*issue, error) {
	var issues []*issue
	for _, entry := range doc.(map[string]interface{})["feed"].(map[string]interface{})["entry"].([]interface{}) {
		issue, err := parseIssueDecodedJson(entry.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

func parseIssues() ([]*issue, error) {
	// TODO: This just pulls in canned test data from one file.
	const filePath = "sample-issues.json"
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return parseIssuesJson(bytes)
}

func main() {
	issues, err := parseIssues()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d issues, from %d to %d\n", len(issues), issues[0].id, issues[len(issues)-1].id)
}
