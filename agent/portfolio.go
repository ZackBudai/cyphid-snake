package agent

// HeuristicPortfolio represents a collection of weighted heuristics.
type HeuristicPortfolio []weightedHeuristic

// weightedHeuristic represents a heuristic with an associated weight and name.
type weightedHeuristic struct {
	Name   string
	F      HeuristicFunc
	Weight float64
}

// HeuristicFunc is a type that represents a heuristic function.
// It takes a GameSnapshot as input and returns a float64 score.
type HeuristicFunc func(GameSnapshot) float64

func NewPortfolio(heuristics ...weightedHeuristic) HeuristicPortfolio {
	return HeuristicPortfolio(heuristics)
}

func NewHeuristic(weight float64, name string, f HeuristicFunc) weightedHeuristic {
	return weightedHeuristic{
		Name:   name,
		F:      f,
		Weight: weight,
	}
}
