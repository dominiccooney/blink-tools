package ml

import (
	"math/rand"
)

type Distribution struct {
	P []float64
}

type CumulativeDistribution struct {
	P []float64
}

// UniformDistribution returns a new, uniform distribution over nitems items.
func UniformDistribution(nitems int) *Distribution {
	distribution := make([]float64, nitems, nitems)
	for i := range distribution {
		distribution[i] = 1.0 / float64(nitems)
	}
	return &Distribution{distribution}
}

func (d *Distribution) Normalize() {
	sum := 0.0
	for _, v := range d.P {
		sum += v
	}
	for i, _ := range d.P {
		d.P[i] /= sum
	}
}

func CumulativeDistributionOfDistribution(dist *Distribution) *CumulativeDistribution {
	cumulative := make([]float64, len(dist.P), len(dist.P))
	sum := 0.0
	for i, p := range dist.P {
		sum += p
		cumulative[i] = sum
	}
	return &CumulativeDistribution{cumulative}
}

// Sample draws a sample, weighted by a distribution, and returns the
// index of the sample.
func (dist *CumulativeDistribution) Sample(r *rand.Rand) int {
	return search(r.Float64(), dist.P, 0, len(dist.P))
}

// search does a binary search to find the index i such that:
// C_i-1 < sample <= C_i
func search(s float64, cumulative []float64, startInclusive int, endExclusive int) int {
	if startInclusive == endExclusive {
		if !((s <= cumulative[0] && startInclusive == 0) || cumulative[startInclusive-1] < s) {
			panic("Search did not find a valid index. Is the cumulative distribution valid?")
		}
		return startInclusive
	}
	mid := startInclusive + (endExclusive-startInclusive)/2
	if s <= cumulative[mid] {
		return search(s, cumulative, startInclusive, mid)
	}
	if mid == startInclusive {
		mid++
	}
	return search(s, cumulative, mid, endExclusive)
}
