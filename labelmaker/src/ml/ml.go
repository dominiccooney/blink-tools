package ml

import (
	"fmt"
	"math"
	"math/rand"
)

type Distribution struct {
	Cumulative []float64
}

// UniformDistribution returns a new, uniform distribution over nitems items.
func UniformDistribution(nitems int) *Distribution {
	distribution := make([]float64, nitems, nitems)
	for i := range distribution {
		distribution[i] = float64(i+1) * 1.0 / float64(nitems)
	}
	return &Distribution{distribution}
}

// Sample draws a sample, weighted by a distribution, and returns the
// index of the sample.
func (dist *Distribution) Sample(r *rand.Rand) int {
	return dist.search(r.Float64(), 0, len(dist.Cumulative))
}

func (dist *Distribution) P(i int) float64 {
	if i == 0 {
		return dist.Cumulative[0]
	} else {
		return dist.Cumulative[i] - dist.Cumulative[i - 1]
	}
}

// search does a binary search to find the index i such that:
// C_i-1 < sample <= C_i
func (dist *Distribution) search(s float64, startInclusive int, endExclusive int) int {
	if startInclusive == endExclusive {
		if !(s <= dist.Cumulative[0] && (startInclusive == 0 || dist.Cumulative[startInclusive-1] < s)) {
			panic("Search did not find a valid index. Is the cumulative distribution valid?")
		}
		return startInclusive
	}
	mid := startInclusive + (endExclusive-startInclusive)/2
	if s < dist.Cumulative[mid] {
		return dist.search(s, startInclusive, mid)
	}
	return dist.search(s, mid, endExclusive)
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

type DecisionStump struct {
	Feature          Feature
	c map[Label]float64
}

func (d *DecisionStump) Predict(e Example, l Label) float64 {
	if !d.Feature.Test(e) {
		return 0.0
	}
	return d.c[l]
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

type key struct {
	b bool
	f Feature
	l Label
}

func (k key) String() string {
	return fmt.Sprintf("{%s %s: %v}", k.f, k.l, k.b)
}


func (stumper *DecisionStumper) NewStump(ds map[Label]*Distribution) *DecisionStump {
	// See Boosting p. 314
	// Pick a feature split that minimizes:
	// Z = 2 * sum: forall values j . forall labels l . sqrt (W_+^jl * W_-^jl)
	// Where:
	// W_b^jl is forall examples i . D(i,l) * 1{x_i is in Xj, Y_i[l]=b}
	// D(i,l) is a distribution for label l over examples i.
	//
	// For now, we only support binary features and stumps which
	// split on one feature.

	// Compute W_+^jl, W_-^jl
	var w map[key]float64 = make(map[key]float64)
	for _, feature := range stumper.features {
		for label, _ := range stumper.labels {
			w[key{true, feature, label}] = 0.0
			w[key{false, feature, label}] = 0.0
		}

		for i, example := range stumper.examples {
			if !feature.Test(example) {
				continue
			}
			for label, _ := range stumper.labels {
				b := example.HasLabel(label)
				w[key{b, feature, label}] += ds[label].P(i)
			}
		}
	}

	for key, val := range w {
		fmt.Printf("%v: %v\n", key, val)
	}

	// Find the feature that minimizes Z_t:
	// TODO: Boosting sums over features in Z_t; should we be seleting the
	// feature with the highest score here to minimize Z_t or the minimum?
	// Doesn't make sense to take something with a high score because
	// that feature sucks, right?
	var fMin Feature = nil
	var zMin float64 = math.MaxFloat64
	for _, feature := range stumper.features {
		zFeature := 0.0
		for label, _ := range stumper.labels {
			zFeature += math.Sqrt(w[key{true, feature, label}] * w[key{false, feature, label}])
		}
		fmt.Printf("%v Z=%f\n", feature, zFeature)
		if zFeature < zMin {
			fMin = feature
			zMin = zFeature
		}
	}

	// Compute c_jl for this feature (j) for each label:
	c := make(map[Label]float64)
	for label, _ := range stumper.labels {
		// 1.0+ is to avoid the case when either of these is 0.
		c[label] = 0.5 * math.Log((1.0 + w[key{true, fMin, label}]) / (1.0 + w[key{false, fMin, label}]))
	}

	return &DecisionStump{fMin, c}
}
