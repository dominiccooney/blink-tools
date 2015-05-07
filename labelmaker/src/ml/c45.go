package ml

import (
	"math"
)

type FeatureNode struct {
	feature  Feature
	positive Classifier
	negative Classifier
}

type LeafNode struct {
	class bool
}

type DecisionTreeBuilder struct {
	features []Feature
	maxDepth int
}

func NewDecisionTreeBuilder(fs []Feature, maxDepth int) *DecisionTreeBuilder {
	return &DecisionTreeBuilder{fs, maxDepth}
}

func (tb *DecisionTreeBuilder) NewClassifier(examples []Example) Classifier {
	return tb.build(1, examples)
}

func info(pos int, neg int) float64 {
	tot := float64(pos + neg)
	pPos := float64(pos) / tot
	pNeg := float64(neg) / tot
	return -pPos*math.Log2(pPos) - pNeg*math.Log2(pNeg)
}

type countKey struct {
	feature bool
	class   Label
}

func (tb *DecisionTreeBuilder) build(depth int, examples []Example) Classifier {
	npos, nneg := 0, 0
	for _, example := range examples {
		if example.Label() {
			npos += 1
		} else {
			nneg += 1
		}
	}

	// If all examples have the same class, predict that class
	if npos == 0 {
		return &LeafNode{false}
	} else if nneg == 0 {
		return &LeafNode{true}
	}

	// Don't build the tree past a certain depth, just predict the
	// prevalent class.
	if depth == tb.maxDepth {
		if npos > nneg {
			return &LeafNode{true}
		} else {
			return &LeafNode{false}
		}
	}

	currentInfo := info(npos, nneg)
	var bestFeature Feature
	bestGain := 0.0

	// Find feature with best information gain. We don't do normalization
	// because all of our features are currently binary valued.
	for _, feature := range tb.features {
		counts := make(map[countKey]int)
		for _, example := range examples {
			f := !math.Signbit(feature.Predict(example))
			counts[countKey{f, example.Label()}]++
		}
		nfeaturePos := counts[countKey{true, true}] + counts[countKey{true, false}]
		pfeaturePos := float64(nfeaturePos) / float64(len(examples))
		infoThisFeature := pfeaturePos*info(counts[countKey{true, true}], counts[countKey{true, false}]) + (1.0-pfeaturePos)*info(counts[countKey{false, true}], counts[countKey{false, false}])

		gain := currentInfo - infoThisFeature
		if gain > bestGain {
			// fmt.Printf("new best feature: %s, gain: %f\n", feature, gain)
			bestGain = gain
			bestFeature = feature
		}
	}

	if bestFeature == nil {
		// FIXME: I'm encoding the base rate here; when does this happen?
		// fmt.Printf("bailing out\n")
		return &LeafNode{false}
	}

	// Split examples into positive and negative for this feature.
	var pos, neg []Example
	for _, example := range examples {
		if math.Signbit(bestFeature.Predict(example)) {
			neg = append(neg, example)
		} else {
			pos = append(pos, example)
		}
	}
	return &FeatureNode{bestFeature, tb.build(depth+1, pos), tb.build(depth+1, neg)}
}

func (n *LeafNode) Predict(e Example) float64 {
	if n.class {
		return 1.0
	} else {
		return -1.0
	}
}

func (n *FeatureNode) Predict(e Example) float64 {
	if math.Signbit(n.feature.Predict(e)) {
		return n.negative.Predict(e)
	} else {
		return n.positive.Predict(e)
	}
}
