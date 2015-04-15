// Machine larnin' on the Chromium Issue tracker.

package main

import (
	"fmt"
	"io/ioutil"
	"issues"
	"path/filepath"
)

func loadIssues() ([]*issues.Issue, error) {
	corpus, err := filepath.Glob("closed-issues-with-cr-label-*.json")
	if err != nil {
		return nil, err
	}

	var is []*issues.Issue = nil
	for _, filePath := range corpus {
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("Reading %s: %v", filePath, err)
		}

		moreIssues, err := issues.ParseIssuesJson(bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing %s: %v", filePath, err)
		}

		is = append(is, moreIssues...)
	}

	return is, nil
}

func main() {
	is, err := loadIssues()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d issues, from %d to %d\n", len(is), is[0].Id, is[len(is)-1].Id)
}
