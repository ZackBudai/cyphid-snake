package agent

import (
	"github.com/Battle-Bunker/cyphid-snake/lib"
	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/client"
	// "github.com/samber/mo"
	"github.com/samber/lo"
	"log"
	"math"
)

// Update the SnakeAgent structure to include SnakeMetadataResponse
type SnakeAgent struct {
	Portfolio HeuristicPortfolio
	Metadata  client.SnakeMetadataResponse
}

func NewSnakeAgent(portfolio HeuristicPortfolio, metadata client.SnakeMetadataResponse) *SnakeAgent {
	return &SnakeAgent{
		Portfolio: portfolio,
		Metadata:  metadata,
	}
}

func (sa *SnakeAgent) ChooseMove(snapshot GameSnapshot) client.MoveResponse {
	you := snapshot.You()
	forwardMoves := you.ForwardMoves()

	// map: move -> set(state snapshots)
	nextStatesMap := make(map[string][]GameSnapshot)
	for _, move := range forwardMoves {
		nextStatesMap[move.Move] = sa.generateNextStates(snapshot, move.Move)
	}

	// map: move -> aggScore
	aggMoveScores := make(map[string]float64)
	for _, heuristic := range sa.Portfolio {
		marginalScores := sa.marginalScoresForHeuristic(heuristic, nextStatesMap)
		for move, score := range marginalScores {
			aggMoveScores[move] += score * heuristic.Weight
		}
	}
	// slice of scores aligned with forwardMoves
	marginalScores := lo.Map(forwardMoves, func(move rules.SnakeMove, _ int) float64 {
		return aggMoveScores[move.Move]
	})

	chosenMove := forwardMoves[lib.SoftmaxSample(marginalScores)]
	return client.MoveResponse{
		Move:  chosenMove.Move,
		Shout: "I'm moving " + chosenMove.Move,
	}
}

func (sa *SnakeAgent) marginalScoresForHeuristic(heuristic weightedHeuristic, nextStatesMap map[string][]GameSnapshot) map[string]float64 {
	moveScores := make(map[string]float64)
	for move, states := range nextStatesMap {
		moveScores[move] = lo.MeanBy(states, heuristic.F)
	}

	log.Printf("    MoveScores for %15s: %+v", heuristic.Name, moveScores)
	meanScore := lo.Mean(lo.Values(moveScores))
	marginalScores := make(map[string]float64)
	for move, score := range moveScores {
		marginalScores[move] = score - meanScore
	}

	roundedMarginalScores := lo.MapValues(marginalScores, func(score float64, _ string) float64 {
		return math.Round(score*100) / 100
	})
	log.Printf("MarginalScores for %15s: %+v", heuristic.Name, roundedMarginalScores)
	return marginalScores
}

func (sa *SnakeAgent) generateNextStates(snapshot GameSnapshot, move string) []GameSnapshot {
	var nextStates []GameSnapshot
	yourID := snapshot.You().ID()

	// Generate all possible move combinations for other snakes
	presetMoves := map[string]rules.SnakeMove{yourID: {ID: yourID, Move: move}}
	moveCombinations := generateForwardMoveCombinations(snapshot.Snakes(), presetMoves)

	// log.Printf("Trying move %s, combinations: %v", move, getMoveComboList(moveCombinations))

	for _, combination := range moveCombinations {
		// Convert the combination map to a slice
		var moveSlice []rules.SnakeMove
		for _, m := range combination {
			moveSlice = append(moveSlice, m)
		}

		if snapshot == nil {
			log.Fatalf("Snapshot is nil before applying moves")
		}
		nextState, err := snapshot.ApplyMoves(moveSlice)

		if err != nil {
			log.Fatalf("Error applying moves: %v", err)
		} else { // Debug the state after ApplyMoves call
			// log.Printf("Next state after applying move: %+v", nextState)
		}
		if nextState != nil {
			nextStates = append(nextStates, nextState)
		}
	}
	// log.Printf("Generated next states: %+v", nextStates)

	return nextStates
}

func generateForwardMoveCombinations(snakes []SnakeSnapshot, presetMoves map[string]rules.SnakeMove) []map[string]rules.SnakeMove {
	presetSnakeIDs := lo.Keys(presetMoves)

	nonPresetSnakes := lo.Filter(snakes, func(snake SnakeSnapshot, _ int) bool {
		return !lo.Contains(presetSnakeIDs, snake.ID())
	})

	// If there are no non-preset snakes, return just our preset combination
	if len(nonPresetSnakes) == 0 {
		return []map[string]rules.SnakeMove{presetMoves}
	}

	nonPresetMoves := lo.Map(nonPresetSnakes, func(snake SnakeSnapshot, _ int) []rules.SnakeMove {
		return snake.ForwardMoves()
	})

	moveCombinations := lib.CartesianProduct(nonPresetMoves...)

	// mix in preset moves to each combo and convert to map from snakeID->move
	mappedCombinations := make([]map[string]rules.SnakeMove, len(moveCombinations))
	for moveSet := range moveCombinations {
		combination := lo.Assign(presetMoves)
		for j, move := range moveSet {
			combination[nonPresetSnakes[j].ID()] = move
		}
		mappedCombinations = append(mappedCombinations, combination)
	}

	return mappedCombinations
}

// for convenient debug printing of move combo collection
func getMoveComboList(moveCombinations []map[string]rules.SnakeMove) [][]string {
	var result [][]string
	for _, combo := range moveCombinations {
		var moves []string
		for _, snakeMove := range combo {
			moves = append(moves, snakeMove.Move)
		}
		result = append(result, moves)
	}
	return result
}
