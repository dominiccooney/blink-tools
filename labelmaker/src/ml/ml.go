package ml

import (
	"fmt"
	"math"
)

type Label bool

type Example interface {
	Label() Label
}

type Feature interface {
	// String returns a human-readable description of the feature.
	String() string
	Predict(Example) float64
}

type andFeature struct {
	f1 Feature
	f2 Feature
}

func (f *andFeature) String() string {
	return fmt.Sprintf("%s && %s", f.f1, f.f2)
}

func (f *andFeature) Predict(e Example) float64 {
	if !math.Signbit(f.f1.Predict(e)) && !math.Signbit(f.f2.Predict(e)) {
		return 1.0
	} else {
		return -1.0
	}
}

type FeatureNegater struct {
	Feature Feature
}

func (f *FeatureNegater) String() string {
	return fmt.Sprintf("not(%s)", f.Feature)
}

func (f *FeatureNegater) Predict(e Example) float64 {
	return -f.Feature.Predict(e)
}
