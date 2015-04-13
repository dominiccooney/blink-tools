// Machine larnin' on the Chromium issue tracker.

package main

import (
	"strconv"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type issue struct {
	id      int
	title   string
	content string
	state   state
	status  status
	labels  []string
}

func parseIssue(entry map[string]interface{}) (*issue, error) {
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
	return "foo"
}

func (p *issueParser) content() string {
	return "bar"
}

func (p *issueParser) state() state {
	return false
}

func (p *issueParser) status() status {
	return statusArchived
}

func (p *issueParser) labels() []string {
	return nil
}

func parseIssues() ([]*issue, error) {
	// TODO: This just pulls in canned test data from one file.
	const filePath = "sample-issues.json"
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

        var doc interface{}
	err = json.Unmarshal(bytes, &doc)
	if err != nil {
		return nil, err
	}

	var issues []*issue
	for _, entry := range doc.(map[string]interface{})["feed"].(map[string]interface{})["entry"].([]interface{}) {
		switch entry.(type) {
		case map[string]interface{}:
			issue, err := parseIssue(entry.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			issues = append(issues, issue)
			break
		default:
			return nil, fmt.Errorf("Entries contained non-entry, %v", entry)
		}
	}
	return issues, nil
}

func main() {
	issues, err := parseIssues()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d issues, from %d to %d\n", len(issues), issues[0].id, issues[len(issues)-1].id)
}
