package ml

import (
	"math/rand"
)

type Sampler struct {
	Nitems                 int
	CumulativeDistribution []float64
	Rand                   *rand.Rand
}

// Sample draws a sample from 0..Nitems-1 according to the
// distribution CumulativeDistribution.
func (sampler *Sampler) Sample() int {
	return sampler.search(sampler.Rand.Float64(), 0, len(sampler.CumulativeDistribution))
}

// search does a binary search to find the index i such that:
// C_i-1 < sample <= C_i
func (sampler *Sampler) search(s float64, startInclusive int, endExclusive int) int {
	if startInclusive == endExclusive {
		if !(s <= sampler.CumulativeDistribution[0] && (startInclusive == 0 || sampler.CumulativeDistribution[startInclusive-1] < s)) {
			panic("Search did not find a valid index. Is the cumulative distribution valid?")
		}
		return startInclusive
	}
	mid := startInclusive + (endExclusive-startInclusive)/2
	if s < sampler.CumulativeDistribution[mid] {
		return sampler.search(s, startInclusive, mid)
	}
	return sampler.search(s, mid, endExclusive)
}
