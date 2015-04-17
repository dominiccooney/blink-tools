// Machine larnin' on the Chromium Issue tracker.

package main

import (
	"fmt"
	"io/ioutil"
	"issues"
	"math/rand"
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

	// Divide into dev, validation and test sets. Use a fixed seed
	// so that the sets are always the same.
	r := rand.New(rand.NewSource(42))
	var dev []*issues.Issue = nil
	var validation []*issues.Issue = nil
	var test []*issues.Issue = nil
	for _, i := range is {
		switch r.Intn(9) {
		case 0:
		case 1:
		case 2:
			test = append(test, i)
			break;
		case 3:
		case 4:
			validation = append(validation, i)
			break;
		default:
			dev = append(dev, i)
			break;
		}
	}

	fmt.Printf("%d issues, from %d to %d\n", len(is), is[0].Id, is[len(is)-1].Id)
}
