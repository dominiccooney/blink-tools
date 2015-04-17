// Machine larnin' on the Chromium Issue tracker.

package main

import (
	"fmt"
	"io/ioutil"
	"issues"
	"math/rand"
	"ml"
	"path/filepath"
	"strings"
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

type IssueExample struct {
	*issues.Issue
}

func (is *IssueExample) Labels() (result []ml.Label) {
	result = nil
	for _, s := range is.IssueLabels {
		if strings.HasPrefix(s, "Cr-") {
			result = append(result, ml.Label(s))
		}
	}
	return
}

func (is *IssueExample) HasLabel(l ml.Label) bool {
	for _, s := range is.IssueLabels {
		if ml.Label(s) == l {
			return true
		}
	}
	return false
}

type titleFeature struct {
	word string
}

func (t *titleFeature) String() string {
	return fmt.Sprintf("title*%s", t.word)
}

func (t *titleFeature) Test(e ml.Example) bool {
	// TODO: Consider testing distinct words because of substring matches.
	return strings.Contains(e.(*IssueExample).Title, t.word)
}

func extractFeatures(examples []ml.Example) (features []ml.Feature) {
	features = nil

	// Titles.
	titleWords := make(map[string]int)
	for _, example := range examples {
		issue := example.(*IssueExample)
		for _, word := range strings.Split(issue.Title, " ") {
			// TODO: Consider lowercasing, cleaning, stemming.
			n, _ := titleWords[word]
			titleWords[word] = n + 1
		}
	}
	ninetyPercentExamples := int(0.9 * float64(len(examples)))
	for word, count := range titleWords {
		if 10 < count && count < ninetyPercentExamples {
			features = append(features, &titleFeature{word})
		}
	}

	return
}

func main() {
	is, err := loadIssues()
	if err != nil {
		panic(err)
	}

	// Divide into dev, validation and test sets. Use a fixed seed
	// so that the sets are always the same.
	r := rand.New(rand.NewSource(42))
	var dev []ml.Example = nil
	var validation []ml.Example = nil
	var test []ml.Example = nil
	for _, i := range is {
		switch r.Intn(9) {
		case 0:
		case 1:
		case 2:
			test = append(test, &IssueExample{i})
			break
		case 3:
		case 4:
			validation = append(validation, &IssueExample{i})
			break
		default:
			dev = append(dev, &IssueExample{i})
			break
		}
	}

	fmt.Printf("%d issues, from %d to %d\n", len(is), is[0].Id, is[len(is)-1].Id)

	// Collect labels.
	labels := make(map[ml.Label]bool)
	for _, is := range dev {
		for _, label := range is.Labels() {
			labels[label] = true
		}
	}

	// Build features.
	features := extractFeatures(dev)
	fmt.Printf("%d features: %v, %v, %v, ...\n", len(features), features[0], features[1], features[2])

	// Build a decision stump.
	dist := make(map[ml.Label]*ml.Distribution)
	for label, _ := range labels {
		dist[label] = ml.UniformDistribution(len(dev))
	}

	stumper := ml.NewDecisionStumper(features, dev)
	booster := ml.NewAdaBoostMH(dev, stumper)


	for i := 0; i < 10; i++ {
		booster.Round()
		fmt.Printf("%d: dev=%d test=%d %v\n", i, booster.Evaluate(dev), booster.Evaluate(test), booster.H[i].Feature)
	}
}
