package main

import (
	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/client"
	"log"
	"math/rand"
	"strconv"
)

func main() {
	RunServer()
}

func info() SnakeMetadataResponse {
	log.Println("INFO")

	return SnakeMetadataResponse{
		APIVersion: "1",
		Author:     "zuthan",
		Color:      "#FF7F7F",
		Head:       "evil",
		Tail:       "nr-booster",
	}
}

func start(_ SnakeRequest) {
	log.Println("GAME START")
}

func end(_ SnakeRequest) {
	log.Printf("GAME OVER\n\n")
}

func move(request SnakeRequest) MoveResponse {
	// Initialize moveSet as empty
	var moveSet []rules.SnakeMove

	// Convert GameState to BoardState
	boardState := request.ConvertToBoardState()

	// Create a new ruleset builder
	builder := rules.NewRulesetBuilder()

	// Configure the builder based on the GameState's ruleset settings
	builder.WithParams(map[string]string{
		"foodSpawnChance":     strconv.Itoa(request.Game.Ruleset.Settings.FoodSpawnChance),
		"minimumFood":         strconv.Itoa(request.Game.Ruleset.Settings.MinimumFood),
		"hazardDamagePerTurn": strconv.Itoa(request.Game.Ruleset.Settings.HazardDamagePerTurn),
	})

	// Construct the ruleset specified in the GameState
	ruleset := builder.NamedRuleset(request.Game.Ruleset.Name)

	// Calculate the next board state
	_, nextState, err := ruleset.Execute(boardState, moveSet)
	if err != nil {
		// Handle error appropriately
		// For now, we will just return a default move
		return MoveResponse{
			Move:  "up",
			Shout: "Error in ruleset execution!",
		}
	}

	println(nextState)

	// Generate a random move for now
	moves := []string{"up", "down", "left", "right"}
	randomMove := moves[rand.Intn(len(moves))]

	// Return the move response
	return MoveResponse{
		Move:  randomMove,
		Shout: "Random move!",
	}
}

func isTeammate(snake1, snake2 client.Snake) bool {
	// In a real implementation, you'd need a way to identify teammates.
	// For this example, we'll consider snakes with the same color as teammates.
	return snake1.Customizations.Color == snake2.Customizations.Color
}
