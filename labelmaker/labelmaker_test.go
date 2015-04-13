package main

import (
	"testing"
)

func TestIssueParserDecodesHtmlContent(t *testing.T) {
	p := newIssueParser(map[string]interface{}{
		"content": map[string]interface{} {
			"$t": "escaped &amp; stuff",
		},
	})
	e := "escaped & stuff"
	s := p.content()
	if e != s {
		t.Errorf("content should be unescaped; expected \"%s\" but got \"%s\"", e, s)
	}
}

