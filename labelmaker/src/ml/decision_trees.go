package ml

import (
	"fmt"
	"math/rand"
)

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
			bestStump = stump
		}
	}

	fmt.Printf("Best stump %f: \"%s\"\n", bestStump.e_t, bestStump.Feature)
	return bestStump
}
