// Fetches issues as JSON from the Chromium issue tracker and saves
// them to sharded files.
//
// See src/issues/issues.go for query string parameters.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

var numFiles = flag.Int("num-files", 1, "number of pages of issues to save")

func optionalStartIndexParameter(startIndex int) string {
	if startIndex == 0 {
		return ""
	} else {
		return fmt.Sprintf("&start-index=%d", startIndex)
	}
}

func main() {
	flag.Parse()

	urlEncodedQuery := "is:closed+cr-"
	issuesPerPage := 1000

	for i := 0; i < *numFiles; i++ {
		url := fmt.Sprintf("https://code.google.com/feeds/issues/p/chromium/issues/full?q=%s&alt=json&max-results=%d%s", urlEncodedQuery, issuesPerPage, optionalStartIndexParameter(i * issuesPerPage))

		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(fmt.Sprintf("closed-issues-with-cr-label-%03d.json", i), body, 0644)
		if err != nil {
			panic(err)
		}

		fmt.Printf(".")
	}
	fmt.Printf("\n")
}
