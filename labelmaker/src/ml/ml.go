package ml

import (
	"fmt"
	"math"
	"math/rand"
)

type Distribution struct {
	P []float64
}

type CumulativeDistribution struct {
	P []float64
}

// UniformDistribution returns a new, uniform distribution over nitems items.
func UniformDistribution(nitems int) *Distribution {
	distribution := make([]float64, nitems, nitems)
	for i := range distribution {
		distribution[i] = 1.0 / float64(nitems)
	}
	return &Distribution{distribution}
}

func (d *Distribution) Normalize() {
	sum := 0.0
	for _, v := range d.P {
		sum += v
	}
	for i, _ := range d.P {
		d.P[i] /= sum
	}
}

func CumulativeDistributionOfDistribution(dist *Distribution) *CumulativeDistribution {
	cumulative := make([]float64, len(dist.P), len(dist.P))
	sum := 0.0
	for i, p := range dist.P {
		sum += p
		cumulative[i] = sum
	}
	return &CumulativeDistribution{cumulative}
}

// Sample draws a sample, weighted by a distribution, and returns the
// index of the sample.
func (dist *CumulativeDistribution) Sample(r *rand.Rand) int {
	return search(r.Float64(), dist.P, 0, len(dist.P))
}

// search does a binary search to find the index i such that:
// C_i-1 < sample <= C_i
func search(s float64, cumulative []float64, startInclusive int, endExclusive int) int {
	if startInclusive == endExclusive {
		if !(s <= cumulative[0] && (startInclusive == 0 || cumulative[startInclusive-1] < s)) {
			panic("Search did not find a valid index. Is the cumulative distribution valid?")
		}
		return startInclusive
	}
	mid := startInclusive + (endExclusive-startInclusive)/2
	if s < cumulative[mid] {
		return search(s, cumulative, startInclusive, mid)
	}
	return search(s, cumulative, mid, endExclusive)
}

type Label bool

type Example interface {
	Label() Label
}

type Feature interface {
	// String returns a human-readable description of the feature.
	String() string
	Test(Example) bool
}

type andFeature struct {
	f1 Feature
	f2 Feature
}

func (f *andFeature) String() string {
	return fmt.Sprintf("%s && %s", f.f1, f.f2)
}

func (f *andFeature) Test(e Example) bool {
	return f.f1.Test(e) && f.f2.Test(e)
}

type FeatureNegater struct {
	Feature Feature
}

func (f *FeatureNegater) String() string {
	return fmt.Sprintf("not(%s)", f.Feature)
}

func (f *FeatureNegater) Test(e Example) bool {
	return !f.Feature.Test(e)
}

type DecisionStump struct {
	Feature Feature
	e_t     float64
}

func (stump *DecisionStump) Predict(e Example) float64 {
	if stump.Feature.Test(e) {
		return 1.0
	} else {
		return -1.0
	}
}

type DecisionStumper struct {
	features []Feature
	examples []Example
	r *rand.Rand
}

func NewDecisionStumper(fs []Feature, es []Example) *DecisionStumper {
	return &DecisionStumper{fs, es, rand.New(rand.NewSource(7))}
}

func (stumper *DecisionStumper) NewStump(ds *Distribution) *DecisionStump {
	var bestStump *DecisionStump = nil

	// Consider random pairs of features as stumps.
	for i := 0; i < 1000; i++ {
		f1 := stumper.features[stumper.r.Intn(len(stumper.features))]
		f2 := stumper.features[stumper.r.Intn(len(stumper.features))]
		var feature Feature = &andFeature{f1, f2}

		counts := make(map[bool]float64)
		for i, example := range stumper.examples {
			counts[feature.Test(example) == bool(example.Label())] += ds.P[i]
		}

		denom := counts[true] + counts[false]
		counts[true] /= denom
		counts[false] /= denom

		if counts[true] < counts[false] {
			feature = &FeatureNegater{feature}
			counts[true], counts[false] = counts[false], counts[true]
		}

		stump := &DecisionStump{feature, counts[false]}

		if bestStump == nil || stump.e_t < bestStump.e_t {
			fmt.Printf("New best stump %f: \"%s\"\n", stump.e_t, stump.Feature)
			bestStump = stump
		}
	}

	return bestStump
}

type AdaBoost struct {
	Examples []Example
	// TODO: Generalize DecisionStumper/DecisionStump to any base learner.
	Stumper *DecisionStumper
	D       Distribution
	H       []*DecisionStump
	A       []float64
}

func NewAdaBoost(es []Example, learner *DecisionStumper) *AdaBoost {
	dist := UniformDistribution(len(es))
	return &AdaBoost{
		es,
		learner,
		*dist,
		nil,
		nil,
	}
}

func float64OfLabel(label Label) float64 {
	if label {
		return 1.0
	} else {
		return -1.0
	}
}

func (a *AdaBoost) Round() {
	h := a.Stumper.NewStump(&a.D)
	if h.e_t < 0.0 || h.e_t > 1.0 {
		fmt.Printf("bad error: %f", h.e_t)
	}
	//h.e_t = math.Max(h.e_t, 1.0e-16)
	a_t := 0.5 * math.Log((1-h.e_t)/h.e_t)
	for i, example := range a.Examples {
		a.D.P[i] *= math.Exp(-a_t * float64OfLabel(example.Label()) * h.Predict(example))
	}
	a.D.Normalize()
	a.H = append(a.H, h)
	a.A = append(a.A, a_t)
}

func (a *AdaBoost) Predict(e Example) float64 {
	sum := 0.0
	for i, h := range a.H {
		sum += a.A[i] * h.Predict(e)
	}
	return sum
}

func DebugCharacterizeWeights(name string, ws []float64) {
	min := math.MaxFloat64
	max := 1.0 - math.MaxFloat64
	sum := 0.0
	for _, w := range ws {
		sum += w
		min = math.Min(min, w)
		max = math.Max(max, w)
	}
	fmt.Printf("%s: %d values, min=%f, mean=%f, max=%f\n", name, len(ws), min, sum / float64(len(ws)), max)
}

// Evaluates the classifier on a test set and returns the error rate.
func (a *AdaBoost) Evaluate(test []Example) float64 {
	var scores []float64
	mispredictions := 0
	for _, example := range test {
		score := a.Predict(example)
		if Label(score > 0.0) != example.Label() {
			mispredictions++
		}
		scores = append(scores, score)
	}
	DebugCharacterizeWeights("scores", scores)
	return float64(mispredictions) / float64(len(test))
}
