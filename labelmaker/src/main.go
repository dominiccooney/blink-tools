// Machine larnin' on the Chromium Issue tracker.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"issues"
	"log"
	"math/rand"
	"ml"
	"os"
	"path/filepath"
	"runtime/pprof"
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

func (is *IssueExample) Label() ml.Label {
	_, ok := is.IssueLabels["Cr-Blink"]
	return ml.Label(ok)
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

func debugCountLabelOccurrence(name string, set []ml.Example) {
	n := 0
	for _, example := range set {
		if example.Label() {
			n++
		}
	}
	fmt.Printf("%s: %d (%.2f)\n", name, n, float64(n) / float64(len(set)))
}

var cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

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
	fmt.Printf("Issues with label:\n")
	debugCountLabelOccurrence("dev", dev)
	debugCountLabelOccurrence("test", test)

	// TODO: Remove this. Shrunk to get profiling results.
	//dev = dev[0:1000]
	//test = test[0:1000]

	// Build features.
	features := extractFeatures(dev)
	fmt.Printf("%d features: %v, %v, %v, ...\n", len(features), features[0], features[1], features[2])

	// Build a decision stump.
	stumper := ml.NewDecisionStumper(features, dev)
	booster := ml.NewAdaBoost(dev, stumper)

	for i := 0; i < 100; i++ {
		booster.Round()
		fmt.Printf("%d: dev=%f test=%f a=%f %v\n", i, booster.Evaluate(dev), booster.Evaluate(test), booster.A[i], booster.H[i].Feature)
	}
}
