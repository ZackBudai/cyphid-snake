package main


func main() {

	metadata := SnakeMetadataResponse{
		APIVersion: "1",
		Author:     "zuthan",
		Color:      "#FF7F7F",
		Head:       "evil",
		Tail:       "nr-booster",
	}
	
	portfolio := NewPortfolio(
		WeightedHeuristic{
			Heuristic: HeuristicHealth,
			Weight:    1.0,
		},
	)

	snakeAgent := NewSnakeAgent(portfolio, metadata)
	server := NewServer(snakeAgent)

	server.Start()
}

