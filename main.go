package main

import (
	"github.com/Battle-Bunker/cyphid-snake/agent"
	"github.com/Battle-Bunker/cyphid-snake/server"
	"github.com/BattlesnakeOfficial/rules/client"
)

func main() {

	metadata := client.SnakeMetadataResponse{
		APIVersion: "1",
		Author:     "zuthan",
		Color:      "#FF7F7F",
		Head:       "evil",
		Tail:       "nr-booster",
	}

	portfolio := agent.NewPortfolio(
		agent.NewHeuristic(1.0, "team-health", HeuristicHealth),
	)

	snakeAgent := agent.NewSnakeAgent(portfolio, metadata)
	server := server.NewServer(snakeAgent)

	server.Start()
}
