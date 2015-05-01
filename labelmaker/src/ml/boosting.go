package ml

import (
	"fmt"
	"math"
)

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
