package agent

import (
	"fmt"
)

// HeuristicPortfolio represents a collection of weighted heuristics.
type HeuristicPortfolio []WeightedHeuristic

// WeightedHeuristic is an interface defining the methods to access properties of a weighted heuristic.
type WeightedHeuristic interface {
	Name() string
	F() HeuristicFunc
	Weight() float64
	NameAndWeight() string
}

// HeuristicFunc is a type that represents a heuristic function.
// It takes a GameSnapshot as input and returns a float64 score.
type HeuristicFunc func(GameSnapshot) float64

func NewPortfolio(heuristics ...WeightedHeuristic) HeuristicPortfolio {
	return HeuristicPortfolio(heuristics)
}

func NewHeuristic(weight float64, name string, f HeuristicFunc) WeightedHeuristic {
	return weightedHeuristicImpl{
		name:   name,
		f:      f,
		weight: weight,
	}
}

// weightedHeuristicImpl represents a heuristic with an associated weight and name.
type weightedHeuristicImpl struct {
	name   string
	f      HeuristicFunc
	weight float64
}

func (w weightedHeuristicImpl) Name() string {
	return w.name
}

func (w weightedHeuristicImpl) F() HeuristicFunc {
	return w.f
}

func (w weightedHeuristicImpl) Weight() float64 {
	return w.weight
}

func (w weightedHeuristicImpl) NameAndWeight() string {
	return fmt.Sprintf("%s, w=%.2f", w.name, w.weight)
}
