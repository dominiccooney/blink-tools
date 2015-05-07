// Machine larnin' on the Chromium Issue tracker.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"issues"
	"log"
	"math"
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

type bigramKey string

func makeBigramKey(w1 string, w2 string) bigramKey {
	return bigramKey(fmt.Sprintf("%s,%s", w1, w2))
}

type IssueExample struct {
	*issues.Issue
	titleBigrams   map[bigramKey]bool
	contentBigrams map[bigramKey]bool
}

func bigramHash(s string) map[bigramKey]bool {
	m := make(map[bigramKey]bool)
	words := strings.Split(s, " ")
	for i := 0; i < len(words)-1; i++ {
		m[makeBigramKey(words[i], words[i+1])] = true
	}
	return m
}

func NewIssueExample(i *issues.Issue) *IssueExample {
	return &IssueExample{i, bigramHash(i.Title), bigramHash(i.Content)}
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

type bigrammer func(e ml.Example) map[bigramKey]bool

type bigram struct {
	label string
	f     bigrammer
	pair  bigramKey
}

func (b *bigram) String() string {
	return fmt.Sprintf("%s*(%s)", b.label, b.pair)
}

func (b *bigram) Predict(e ml.Example) float64 {
	if _, ok := b.f(e)[b.pair]; ok {
		return 1.0
	}
	return -1.0
}

func extractBigrams(label string, f bigrammer, e ml.Example) []ml.Feature {
	bis := f(e)
	grams := make([]ml.Feature, len(bis), len(bis))
	i := 0
	for b := range bis {
		grams[i] = &bigram{label, f, b}
		i++
	}
	return grams
}

func extractFeatures(examples []ml.Example) (features []ml.Feature) {
	features = nil

	var title bigrammer = func(e ml.Example) map[bigramKey]bool {
		return (e.(*IssueExample)).titleBigrams
	}
	var content bigrammer = func(e ml.Example) map[bigramKey]bool {
		return (e.(*IssueExample)).contentBigrams
	}
	bigrams := make(map[string]ml.Feature)
	for _, example := range examples {
		for _, gram := range extractBigrams("title", title, example) {
			bigrams[gram.String()] = gram
		}
		for _, gram := range extractBigrams("content", content, example) {
			bigrams[gram.String()] = gram
		}
	}

	minExamples := int(0.001 * float64(len(examples)))
	maxExamples := int(0.95 * float64(len(examples)))
	for _, gram := range bigrams {
		count := 0
		for _, example := range examples {
			if !math.Signbit(gram.Predict(example)) {
				count++
			}
		}

		if minExamples <= count && count <= maxExamples {
			features = append(features, gram)
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
	fmt.Printf("%s: %d (%.2f)\n", name, n, float64(n)/float64(len(set)))
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

func trimToBalanceClasses(r *rand.Rand, es []ml.Example) []ml.Example {
	var xs []ml.Example
	var ys []ml.Example
	for i, e := range es {
		if e.Label() {
			xs = append(xs, e)
		} else if r.Intn(i + 1) < len(xs) {
			ys = append(ys, e)
		}
	}
	return append(xs, ys...)
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
			test = append(test, NewIssueExample(i))
			break
		case 3:
		case 4:
			validation = append(validation, NewIssueExample(i))
			break
		default:
			dev = append(dev, NewIssueExample(i))
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
	dev = trimToBalanceClasses(r, dev)
	debugCountLabelOccurrence("trimmed dev", dev)

	// Build features.
	features := extractFeatures(dev)
	fmt.Printf("%d features: %v, %v, %v, ...\n", len(features), features[0], features[1], features[2])

	// Build a decision tree.
	// stumper := ml.NewDecisionStumper(features, dev, r)
	maxDecisionTreeDepth := 5
	treeBuilder := ml.NewDecisionTreeBuilder(features, maxDecisionTreeDepth)
	booster := ml.NewAdaBoost(dev, treeBuilder, r)

	for i := 0; i < 1000; i++ {
		booster.Round(100)
		fmt.Printf("%d: dev=%f test=%f a=%f\n", i, booster.Evaluate(dev), booster.Evaluate(test), booster.A[i])
		debugDumpExampleWeights(booster)
	}
}
