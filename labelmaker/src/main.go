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

func loadIssues(glob string) ([]*issues.Issue, error) {
	corpus, err := filepath.Glob(glob)
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

func (t *titleFeature) Predict(e ml.Example) float64 {
	// TODO: Consider testing distinct words because of substring matches.
	if strings.Contains(e.(*IssueExample).Title, t.word) {
		return 1.0
	} else {
		return -1.0
	}
}

type contentFeature struct {
	word string
}

func (f *contentFeature) String() string {
	return f.word
}

func (f *contentFeature) Predict(e ml.Example) float64 {
	// TODO: Consider testing distinct words because of substring matches.
	if strings.Contains(e.(*IssueExample).Content, f.word) {
		return 1.0
	} else {
		return -1.0
	}
}

func extractFeatures(examples []ml.Example) (features []ml.Feature) {
	features = nil

	// FIXME: There's a lot of duplication here with how title and
	// body are handled.
	titleWords := make(map[string]int)
	bodyWords := make(map[string]int)
	for _, example := range examples {
		issue := example.(*IssueExample)
		for _, word := range strings.Split(issue.Title, " ") {
			// TODO: Consider lowercasing, cleaning, stemming.
			n, _ := titleWords[word]
			titleWords[word] = n + 1
		}
		for _, word := range strings.Split(issue.Content, " ") {
			n, _ := bodyWords[word]
			bodyWords[word] = n + 1
		}
	}
	minExamples := int(0.01 * float64(len(examples)))
	maxExamples := int(0.50 * float64(len(examples)))
	for word, count := range titleWords {
		if word == "" {
			continue
		}

		if minExamples <= count && count <= maxExamples {
			features = append(features, &titleFeature{word})
		}
	}
	for word, count := range bodyWords {
		if word == "" {
			continue
		}

		if minExamples <= count && count <= maxExamples {
			features = append(features, &contentFeature{word})
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

func debugDumpExampleWeights(a *ml.AdaBoost) {
	var positives []float64
	var negatives []float64
	for i, example := range a.Examples {
		if example.Label() {
			positives = append(positives, a.D.P[i])
		} else {
			negatives = append(negatives, a.D.P[i])
		}
	}
	ml.DebugCharacterizeWeights("+ve", positives)
	ml.DebugCharacterizeWeights("-ve", negatives)
}

var cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file")
var dataset = flag.String("dataset", "small", "which dataset to use (small, large)")

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

	is, err := loadIssues(fmt.Sprintf("../datasets/%s/closed-issues-with-cr-label-*.json", *dataset))
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
	dev = dev[0:1000]
	test = test[0:1000]

	// Build features.
	features := extractFeatures(dev)
	fmt.Printf("%d features: %v, %v, %v, ...\n", len(features), features[0], features[1], features[2])

	// Build a decision tree.
	// stumper := ml.NewDecisionStumper(features, dev, r)
	maxDecisionTreeDepth := 4
	treeBuilder := ml.NewDecisionTreeBuilder(features, maxDecisionTreeDepth)
	booster := ml.NewAdaBoost(dev, treeBuilder, r)

	for i := 0; i < 1000; i++ {
		booster.Round(1000)
		fmt.Printf("%d: dev=%f test=%f a=%f\n", i, booster.Evaluate(dev), booster.Evaluate(test), booster.A[i])
		debugDumpExampleWeights(booster)
	}
}
