package issues

import (
	"testing"
)

func (xs Labels) equals(ys Labels) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if xs[i] != ys[i] {
			return false
		}
	}
	return true
}

func (i Issue) equals(j Issue) bool {
	return i.Id == j.Id && i.Title == j.Title && i.Content == j.Content && i.State == j.State && i.Status == j.Status && i.Labels.equals(j.Labels)
}

func TestParseIssues(t *testing.T) {
	b := []byte(jsonIssuesDoc)
	issues, err := ParseIssuesJson(b)
	if err != nil {
		t.Errorf("should have parsed JSON issues successfully: %v", err)
		return
	}
	if 2 != len(issues) {
		t.Errorf("expected to parse 2 issues but was %d", len(issues))
	}
	expected := Issue{
		476406,
		"Title of the first issue",
		"The < content of the first issue",
		StateClosed,
		StatusWontFix,
		[]string{"OS-Mac", "Pri-2", "Type-Bug", "OS-Linux", "clang"},
	}
	if !expected.equals(*issues[0]) {
		t.Errorf("expected the first issue to be %v but was %v", expected, *issues[0])
	}
}

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

func TestIssueParserEmptyLabels(t *testing.T) {
	b := []byte(jsonIssueWithNoLabelsDoc)
	issues, err := ParseIssuesJson(b)
	if err != nil {
		t.Errorf("should have parsed JSON issues successfully: %v", err)
		return
	}
	if 1 != len(issues) {
		t.Errorf("expected to parse 1 issue but was %d", len(issues))
	}
	if nil != issues[0].Labels {
		t.Errorf("expected the issue to have no labels but was %v", issues[0].Labels)
	}
}

const jsonIssuesDoc = `{"version":"1.0","encoding":"UTF-8","feed":{"xmlns":"http://www.w3.org/2005/Atom","xmlns$openSearch":"http://a9.com/-/spec/opensearch/1.1/","xmlns$gd":"http://schemas.google.com/g/2005","xmlns$issues":"http://schemas.google.com/projecthosting/issues/2009","id":{"$t":"http://code.google.com/feeds/issues/p/chromium/issues/full"},"updated":{"$t":"2015-04-13T05:44:55.600Z"},"title":{"$t":"Issues - chromium"},"subtitle":{"$t":"Issues - chromium"},"link":[{"rel":"alternate","type":"text/html","href":"http://code.google.com/p/chromium/issues/list"},{"rel":"http://schemas.google.com/g/2005#feed","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full"},{"rel":"http://schemas.google.com/g/2005#post","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full"},{"rel":"self","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full?alt=json&q=-is%3Aopen&max-results=100"},{"rel":"next","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full?alt=json&q=-is%3Aopen&start-index=101&max-results=100"}],"generator":{"$t":"ProjectHosting","version":"1.0","uri":"http://code.google.com/feeds/issues"},"openSearch$totalResults":{"$t":272989},"openSearch$startIndex":{"$t":1},"openSearch$itemsPerPage":{"$t":100},"entry":[{"gd$etag":"W/\"D0MHR347eCl7ImA9XRRbGEQ.\"","id":{"$t":"http://code.google.com/feeds/issues/p/chromium/issues/full/476406"},"published":{"$t":"2015-04-13T00:17:39.000Z"},"updated":{"$t":"2015-04-13T03:23:56.000Z"},"title":{"$t":"Title of the first issue"},"content":{"$t":"The &lt; content of the first issue","type":"html"},"link":[{"rel":"replies","type":"application/atom+xml","href":"http://code.google.com/feeds/issues/p/chromium/issues/476406/comments/full"},{"rel":"alternate","type":"text/html","href":"http://code.google.com/p/chromium/issues/detail?id=476406"},{"rel":"self","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full/476406"}],"author":[{"name":{"$t":"author@chromium.org"},"uri":{"$t":"/u/author@chromium.org/"}}],"issues$cc":[{"issues$uri":{"$t":"/u/118337007454936871784/"},"issues$username":{"$t":"h...@chromium.org"}}],"issues$closedDate":{"$t":"2015-04-13T03:23:56.000Z"},"issues$id":{"$t":476406},"issues$label":[{"$t":"OS-Mac"},{"$t":"Pri-2"},{"$t":"Type-Bug"},{"$t":"OS-Linux"},{"$t":"clang"}],"issues$stars":{"$t":1},"issues$state":{"$t":"closed"},"issues$status":{"$t":"WontFix"}},{"gd$etag":"W/\"Dk4BQH47eCl7ImA9XRRbGEg.\"","id":{"$t":"http://code.google.com/feeds/issues/p/chromium/issues/full/476379"},"published":{"$t":"2015-04-12T15:13:42.000Z"},"updated":{"$t":"2015-04-12T16:09:11.000Z"},"title":{"$t":"Title of the second issue"},"content":{"$t":"The content of the second issue","type":"html"},"link":[{"rel":"replies","type":"application/atom+xml","href":"http://code.google.com/feeds/issues/p/chromium/issues/476379/comments/full"},{"rel":"alternate","type":"text/html","href":"http://code.google.com/p/chromium/issues/detail?id=476379"},{"rel":"self","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full/476379"}],"author":[{"name":{"$t":"SunFi...@gmail.com"},"uri":{"$t":"/u/100360195550669844867/"}}],"issues$closedDate":{"$t":"2015-04-12T15:25:17.000Z"},"issues$id":{"$t":476379},"issues$label":[{"$t":"Cr-Platform-DevTools"},{"$t":"Pri-2"},{"$t":"Via-Wizard"},{"$t":"Type-Bug"},{"$t":"OS-Mac"}],"issues$stars":{"$t":1},"issues$state":{"$t":"closed"},"issues$status":{"$t":"WontFix"}}]}}`

const jsonIssueWithNoLabelsDoc = `{"version":"1.0","encoding":"UTF-8","feed":{"xmlns":"http://www.w3.org/2005/Atom","xmlns$openSearch":"http://a9.com/-/spec/opensearch/1.1/","xmlns$gd":"http://schemas.google.com/g/2005","xmlns$issues":"http://schemas.google.com/projecthosting/issues/2009","id":{"$t":"http://code.google.com/feeds/issues/p/chromium/issues/full"},"updated":{"$t":"2015-04-13T05:44:55.600Z"},"title":{"$t":"Issues - chromium"},"subtitle":{"$t":"Issues - chromium"},"link":[{"rel":"alternate","type":"text/html","href":"http://code.google.com/p/chromium/issues/list"},{"rel":"http://schemas.google.com/g/2005#feed","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full"},{"rel":"http://schemas.google.com/g/2005#post","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full"},{"rel":"self","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full?alt=json&q=-is%3Aopen&max-results=100"},{"rel":"next","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full?alt=json&q=-is%3Aopen&start-index=101&max-results=100"}],"generator":{"$t":"ProjectHosting","version":"1.0","uri":"http://code.google.com/feeds/issues"},"openSearch$totalResults":{"$t":272989},"openSearch$startIndex":{"$t":1},"openSearch$itemsPerPage":{"$t":100},"entry":[{"gd$etag":"W/\"D0MHR347eCl7ImA9XRRbGEQ.\"","id":{"$t":"http://code.google.com/feeds/issues/p/chromium/issues/full/476406"},"published":{"$t":"2015-04-13T00:17:39.000Z"},"updated":{"$t":"2015-04-13T03:23:56.000Z"},"title":{"$t":"Title of the first issue"},"content":{"$t":"The &lt; content of the first issue","type":"html"},"link":[{"rel":"replies","type":"application/atom+xml","href":"http://code.google.com/feeds/issues/p/chromium/issues/476406/comments/full"},{"rel":"alternate","type":"text/html","href":"http://code.google.com/p/chromium/issues/detail?id=476406"},{"rel":"self","type":"application/atom+xml","href":"https://code.google.com/feeds/issues/p/chromium/issues/full/476406"}],"author":[{"name":{"$t":"author@chromium.org"},"uri":{"$t":"/u/author@chromium.org/"}}],"issues$cc":[{"issues$uri":{"$t":"/u/118337007454936871784/"},"issues$username":{"$t":"h...@chromium.org"}}],"issues$closedDate":{"$t":"2015-04-13T03:23:56.000Z"},"issues$id":{"$t":476406},"issues$stars":{"$t":1},"issues$state":{"$t":"closed"},"issues$status":{"$t":"WontFix"}}]}}`
