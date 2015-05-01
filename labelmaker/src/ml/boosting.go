package ml

import (
	"fmt"
	"math"
	"math/rand"
)

type Learner interface {
	NewClassifier([]Example) Classifier
}

type Classifier interface {
	// Returns -1.0 for the negative class, and 1.0 for the positive class.
	Predict(Example) float64
}

type AdaBoost struct {
	Examples []Example
	Learner Learner
	D       *Distribution
	H       []Classifier
	A       []float64
	rand    *rand.Rand
}

func NewAdaBoost(es []Example, learner Learner, r *rand.Rand) *AdaBoost {
	dist := UniformDistribution(len(es))
	return &AdaBoost{
		es,
		learner,
		dist,
		nil,
		nil,
		r,
	}
}

func float64OfLabel(label Label) float64 {
	if label {
		return 1.0
	} else {
		return -1.0
	}
}

// Evaluates Classifier c and on examples and returns the error rate, 0.0-1.0.
func evaluateClassifier(c Classifier, examples []Example) float64 {
	return evaluateClassifierWeighted(c, examples, UniformDistribution(len(examples)))
}

func evaluateClassifierWeighted(c Classifier, examples []Example, d *Distribution) float64 {
	misclassifications := 0.0
	for i, example := range examples {
		if !math.Signbit(c.Predict(example)) != bool(example.Label()) {
			misclassifications += d.P[i]
		}
	}
	return misclassifications
}

func (a *AdaBoost) Round(nexamples int) {
	// Sample from the examples for this round.
	cumulative := CumulativeDistributionOfDistribution(a.D)
	var examples []Example
	for i := 0; i < nexamples; i++ {
		examples = append(examples, a.Examples[cumulative.Sample(a.rand)])
	}

	h := a.Learner.NewClassifier(examples)

	// Calculate the error of this classifier.
	e_t := evaluateClassifierWeighted(h, a.Examples, a.D)
	e_t = math.Max(e_t, math.SmallestNonzeroFloat64)
	a_t := 0.5 * math.Log((1-e_t)/e_t)
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
