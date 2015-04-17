package ml

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestSamplerUnitaryDistribution(t *testing.T) {
	d := UniformDistribution(1)
	x := d.Sample(rand.New(rand.NewSource(0)))
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

func (r *reflectedFeature) Test(e Example) bool {
	val := reflect.ValueOf(e).MethodByName(r.name).Call(nil)[0].String()
	return r.value == val
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

func (d *datum) Labels() []Label {
	return []Label{d.class}
}

func (d *datum) HasLabel(l Label) bool {
	return d.class == l
}

func TestDecisionStump(t *testing.T) {
	dist := map[Label]*Distribution{
		"vehicle": UniformDistribution(4),
		"fruit":   UniformDistribution(4),
	}

	dataset := []Example{
		&datum{"red", "heavy", "vehicle"},
		&datum{"red", "light", "fruit"},
		&datum{"yellow", "light", "fruit"},
		&datum{"yellow", "light", "vehicle"},
	}

	features := []Feature{
		&reflectedFeature{"Color", "red"},
		&reflectedFeature{"Color", "yellow"},
		&reflectedFeature{"Weight", "heavy"},
	}

	stumper := NewDecisionStumper(features, dataset)
	stump := stumper.NewStump(dist)
	if stump.Feature != features[2] {
		t.Errorf("expected decision stump to split on Weight*heavy but split on %v", stump.Feature)
	}
}
