package ml

import (
	"math/rand"
	"testing"
)

func TestSamplerUnitaryDistribution(t *testing.T) {
	s := &Sampler{
		1,
		[]float64{1.0},
		rand.New(rand.NewSource(0)),
	}
	x := s.Sample()
	if 0 != x {
		t.Errorf("Sampling the unitary distribution should produce index 0, was %d", x)
	}
}
