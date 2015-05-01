package ml

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestSamplerUnitaryDistribution(t *testing.T) {
	d := UniformDistribution(1)
	c := CumulativeDistributionOfDistribution(d)
	x := c.Sample(rand.New(rand.NewSource(0)))
	if 0 != x {
		t.Errorf("Sampling the unitary distribution should produce index 0, was %d", x)
	}
}

type reflectedFeature struct {
	name  string
	value string
}

func (r *reflectedFeature) String() string {
	return fmt.Sprintf("%s*%v", r.name, r.value)
}

func (r *reflectedFeature) Predict(e Example) float64 {
	val := reflect.ValueOf(e).MethodByName(r.name).Call(nil)[0].String()
	if r.value == val {
		return 1.0
	} else {
		return -1.0
	}
}

type datum struct {
	color  string
	weight string
	class  Label
}

func (d *datum) Color() string {
	return d.color
}

func (d *datum) Weight() string {
	return d.weight
}

func (d *datum) Label() Label {
	return d.class
}

func TestDecisionStump(t *testing.T) {
	dataset := []Example{
		&datum{"red", "heavy", true},
		&datum{"red", "light", false},
		&datum{"yellow", "light", false},
		&datum{"yellow", "light", true},
	}

	features := []Feature{
		&reflectedFeature{"Color", "red"},
		&reflectedFeature{"Color", "yellow"},
		&reflectedFeature{"Weight", "heavy"},
	}

	r := rand.New(rand.NewSource(42))
	stumper := NewDecisionStumper(features, dataset, r)
	stump := stumper.NewClassifier(dataset)
	if stump != features[2] {
		t.Errorf("expected decision stump to split on Weight*heavy but split on %v", stump)
	}
}
