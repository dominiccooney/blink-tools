package ml

import (
	"fmt"
)

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
