package ml

import (
	"fmt"
	"math/rand"
)

type DecisionStumper struct {
	features []Feature
	examples []Example
	r *rand.Rand
}

func NewDecisionStumper(fs []Feature, es []Example, r *rand.Rand) *DecisionStumper {
	return &DecisionStumper{fs, es, r}
}

func (stumper *DecisionStumper) NewClassifier(examples []Example) Classifier {
	var bestStump Feature = nil
	bestError := 1.0

	// Consider random pairs of features as stumps.
	for i := 0; i < 1000; i++ {
		f1 := stumper.features[stumper.r.Intn(len(stumper.features))]
		f2 := stumper.features[stumper.r.Intn(len(stumper.features))]
		var feature Feature = &andFeature{f1, f2}

		error := evaluateClassifier(feature, examples)

		if error > 0.5 {
			feature = &FeatureNegater{feature}
			error = 1.0 - error
		}

		if bestStump == nil || error < bestError {
			bestStump = feature
			bestError = error
		}
	}

	fmt.Printf("Best stump %f: \"%s\"\n", bestError, bestStump)
	return bestStump
}
