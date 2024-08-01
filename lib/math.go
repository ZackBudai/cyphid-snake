package lib

import (
	"math"
	"math/rand"
)

func Softmax(inputs []float64) []float64 {
	var max float64 = inputs[0]
	for _, v := range inputs[1:] {
		if v > max {
			max = v
		}
	}

	exps := make([]float64, len(inputs))
	var sum float64
	for i, v := range inputs {
		exps[i] = math.Exp(v - max)
		sum += exps[i]
	}

	for i := range exps {
		exps[i] /= sum
	}
	return exps
}

func SoftmaxSample(inputs []float64) int {
	probs := Softmax(inputs)
	r := rand.Float64()
	var cumulativeProb float64
	for i, prob := range probs {
		cumulativeProb += prob
		if r <= cumulativeProb {
			return i
		}
	}
	return len(inputs) - 1
}
