package main

import (
	"testing"
)

func TestIssueParserTitle(t *testing.T) {
	e := "Not escaped & stuff"
	p := newIssueParser(map[string]interface{}{
		"title": map[string]interface{}{
			"$t": e,
		},
	})
	s := p.title()
	if e != s {
		t.Errorf("title should be \"%s\" but was \"%s\"", e, s)
	}
}

func TestIssueParserDecodesHtmlContent(t *testing.T) {
	p := newIssueParser(map[string]interface{}{
		"content": map[string]interface{}{
			"$t": "escaped &amp; stuff",
		},
	})
	e := "escaped & stuff"
	s := p.content()
	if e != s {
		t.Errorf("content should be unescaped; expected \"%s\" but was \"%s\"", e, s)
	}
}

