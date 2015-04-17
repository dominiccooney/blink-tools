package ml

import (
	"fmt"
	"math"
	"math/rand"
)

type Distribution struct {
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

// Sample draws a sample, weighted by a distribution, and returns the
// index of the sample.
func (dist *Distribution) Sample(r *rand.Rand) int {
	cumulative := make([]float64, len(dist.P), len(dist.P))
	sum := 0.0
	for i, p := range dist.P {
		sum += p
		cumulative[i] = sum
	}
	return search(r.Float64(), cumulative, 0, len(cumulative))
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

type Label string

type Example interface {
	Labels() []Label
	HasLabel(Label) bool
}

type Feature interface {
	// String returns a human-readable description of the feature.
	String() string
	Test(Example) bool
}

type stumpProbabilityKey struct {
	hasFeature bool
	label Label
	hasLabel bool
}

type DecisionStump struct {
	Feature Feature
	p map[stumpProbabilityKey]float64
	e_t float64
}

// Predict indicates whether e has label l.
func (d *DecisionStump) Predict(e Example, l Label) bool {
	return d.p[stumpProbabilityKey{d.Feature.Test(e), l, true}] > 0.5
}

type DecisionStumper struct {
	labels   map[Label]bool
	features []Feature
	examples []Example
}

func NewDecisionStumper(fs []Feature, es []Example) *DecisionStumper {
	// Collect a set of all labels.
	labels := make(map[Label]bool)
	for _, e := range es {
		for _, l := range e.Labels() {
			labels[l] = true
		}
	}
	return &DecisionStumper{labels, fs, es}
}

func (stumper *DecisionStumper) NewStump(ds map[Label]*Distribution) *DecisionStump {
	var bestStump *DecisionStump = nil

	// Consider each feature as a potential stump.
	for _, feature := range stumper.features {
		counts := make(map[stumpProbabilityKey]float64)
		for i, example := range stumper.examples {
			b := feature.Test(example)
			for label, _ := range stumper.labels {
				counts[stumpProbabilityKey{b, label, example.HasLabel(label)}] += ds[label].P[i];
			}
		}

		// Normalize probabilities.
		for _, hasFeature := range []bool{false, true} {
			for label, _ := range stumper.labels {
				denom := counts[stumpProbabilityKey{hasFeature, label, false}] + counts[stumpProbabilityKey{hasFeature, label, true}]
				counts[stumpProbabilityKey{hasFeature, label, false}] /= denom
				counts[stumpProbabilityKey{hasFeature, label, true}] /= denom
			}
		}

		stump := &DecisionStump{feature, counts, 0.0}

		// Calculate e_t, the Hamming loss splitting on feature:
		for i, example := range stumper.examples {
			for label, _ := range stumper.labels {
				if example.HasLabel(label) != stump.Predict(example, label) {
					stump.e_t += ds[label].P[i];
				}
			}
		}

		if bestStump == nil || stump.e_t < bestStump.e_t {
			fmt.Printf("New best stump %f: \"%s\"\n", stump.e_t, stump.Feature)
			bestStump = stump
		}
	}

	return bestStump
}

type AdaBoostMH struct {
	Examples []Example
	// TODO: Generalize DecisionStumper/DecisionStump to any base learner.
	Stumper *DecisionStumper
	D       map[Label]*Distribution
	H       []*DecisionStump
	A       []float64
}

func NewAdaBoostMH(es []Example, learner *DecisionStumper) *AdaBoostMH {
	dist := make(map[Label]*Distribution)
	for label, _ := range learner.labels {
		dist[label] = UniformDistribution(len(es))

		// The probability distribution is normalized by
		// number of labels.
		for i := range dist[label].P {
			dist[label].P[i] /= float64(len(learner.labels))
		}
	}
	return &AdaBoostMH{
		es,
		learner,
		dist,
		nil,
		nil,
	}
}

func hasLabel(e Example, l Label) float64 {
	if e.HasLabel(l) {
		return 1.0
	} else {
		return -1.0
	}
}

func predict(h *DecisionStump, e Example, l Label) float64 {
	if h.Predict(e, l) {
		return 1.0
	} else {
		return -1.0
	}
}

func (a *AdaBoostMH) Round() {
	h := a.Stumper.NewStump(a.D)
	if h.e_t < 0.0 || h.e_t > 1.0 {
		fmt.Printf("bad error: %f", h.e_t)
	}
	a_t := 0.5 * math.Log((1 - h.e_t) / h.e_t)
	scale := make([]float64, len(a.Examples), len(a.Examples))
	sum := 0.0
	for label, _ := range a.Stumper.labels {
		for i, example := range a.Examples {
			scale[i] = math.Exp(-a_t * hasLabel(example, label)*predict(h, example, label))
			sum += scale[i]
		}
	}
	for label, _ := range a.Stumper.labels {
		for i := range a.Examples {
			a.D[label].P[i] *= scale[i] / sum
		}
	}
	a.H = append(a.H, h)
	a.A = append(a.A, a_t)
}

func (a *AdaBoostMH) Predict(e Example, l Label) float64 {
	sum := 0.0
	for i, h := range a.H {
		sum += a.A[i] * predict(h, e, l)
	}
	//fmt.Printf("%s %f %s\n", fmt.Sprintf("%s", e)[22:40], sum, l)
	return sum
}

// Hamming distance; lower is better.
func (a *AdaBoostMH) Evaluate(test []Example) int {
	dist := 0
	for _, example := range test {
		for label, _ := range a.Stumper.labels {
			if a.Predict(example, label) > 0.0 && example.HasLabel(label) {
				continue
			}
			dist++
		}
	}
	// TODO: Maybe normalize this?
	return dist
}
