package agent

// HeuristicFunc is a type that represents a heuristic function.
// It takes a GameSnapshot as input and returns a float64 score.
type HeuristicFunc func(GameSnapshot) int

type WeightedHeuristic struct {
	Name      string
	Heuristic HeuristicFunc
	Weight    float64
}

type Portfolio []WeightedHeuristic

func NewPortfolio(heuristics ...WeightedHeuristic) Portfolio {
	return Portfolio(heuristics)
}
