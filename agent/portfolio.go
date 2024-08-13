package agent

// HeuristicFunc is a type that represents a heuristic function.
// It takes a GameSnapshot as input and returns a float64 score.
type HeuristicFunc func(GameSnapshot) int

type weightedHeuristic struct {
	Name      string
	Heuristic HeuristicFunc
	Weight    float64
}

type Portfolio []weightedHeuristic

func NewPortfolio(heuristics ...weightedHeuristic) Portfolio {
	return Portfolio(heuristics)
}

func NewHeuristic(weight float64, name string, heuristic HeuristicFunc) weightedHeuristic {
	return weightedHeuristic{
		Name: name,
		Heuristic: heuristic,
		Weight: weight,
	}
}
