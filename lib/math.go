package lib

import (
	"github.com/samber/lo"
	"math"
	"math/rand"
)

func SoftmaxWithTemp(inputs []float64, temp float64) []float64 {
	max := lo.Max(inputs)

	exps := lo.Map(inputs, func(v float64, _ int) float64 {
		return math.Exp((v - max) / temp)
	})
	sum := lo.Sum(exps)

	exps = lo.Map(exps, func(v float64, _ int) float64 {
		return v / sum
	})

	return exps
}

func Softmax(inputs []float64) []float64 {
	return SoftmaxWithTemp(inputs, 1.0)
}

func SampleFromWeights(weights []float64) int {
		r := rand.Float64()
		var cumulativeProb float64
		for i, weight := range weights {
			cumulativeProb += weight
			if r <= cumulativeProb {
				return i
			}
		}
		return len(weights) - 1
}

func SoftmaxSample(inputs []float64) int {
		probs := Softmax(inputs)
		return SampleFromWeights(probs)
}
