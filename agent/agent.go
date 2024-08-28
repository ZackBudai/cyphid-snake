package agent

import (
	"github.com/Battle-Bunker/cyphid-snake/lib"
	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/client"

	// "github.com/samber/mo"
	"fmt"
	"log"
	// "math"
	"slices"
	"strings"

	"github.com/samber/lo"
)

// Update the SnakeAgent structure to include SnakeMetadataResponse
type SnakeAgent struct {
	Portfolio   HeuristicPortfolio
	Temperature float64
	Metadata    client.SnakeMetadataResponse
}

func NewSnakeAgentWithTemp(portfolio HeuristicPortfolio, temperature float64, metadata client.SnakeMetadataResponse) *SnakeAgent {
	return &SnakeAgent{
		Portfolio:   portfolio,
		Temperature: temperature,
		Metadata:    metadata,
	}
}

func NewSnakeAgent(portfolio HeuristicPortfolio, metadata client.SnakeMetadataResponse) *SnakeAgent {
	return &SnakeAgent{
		Portfolio:   portfolio,
		Temperature: 5.0,
		Metadata:    metadata,
	}
}

func (sa *SnakeAgent) ChooseMove(snapshot GameSnapshot) client.MoveResponse {
	you := snapshot.You()
	forwardMoves := you.ForwardMoves()

	forwardMoveStrs := lo.Map(forwardMoves, func(move rules.SnakeMove, _ int) string { return move.Move })
	slices.Sort(forwardMoveStrs)
	log.Printf("\n\n ### Start Turn %d: Forward Moves = %v", snapshot.Turn(), forwardMoveStrs)

	// map: move -> set(state snapshots)
	nextStatesMap := make(map[string][]GameSnapshot)
	for _, move := range forwardMoveStrs {
		nextStatesMap[move] = sa.generateNextStates(snapshot, move)
	}

	// slice of maps, for each heuristic, giving mapping: move -> aggScore
	heuristicScores := lo.Map(sa.Portfolio, func(heuristic WeightedHeuristic, _ int) map[string]float64 {
		return sa.weightedScoresForHeuristic(heuristic, nextStatesMap, forwardMoveStrs)
	})

	totalHeuristicWeight := lo.SumBy(sa.Portfolio, func(heuristic WeightedHeuristic) float64 {
		return heuristic.Weight()
	})

	// slice of scores aligned with forwardMoveStrs
	normalizedScores := lo.Map(forwardMoveStrs, func(move string, _ int) float64 {
		return lo.SumBy(heuristicScores, func(scores map[string]float64) float64 {
			return scores[move] / totalHeuristicWeight
		})
	})

	probs := lib.SoftmaxWithTemp(normalizedScores, sa.Temperature)

	log.Printf("### %36s: %s", "Aggregate move weights", strings.Join(lo.Map(forwardMoveStrs, func(move string, i int) string {
		return fmt.Sprintf("%s=%6.1f", move, normalizedScores[i])
	}), ", "))
	log.Printf("### %36s: %s", "Aggregate move probabilities", strings.Join(lo.Map(forwardMoveStrs, func(move string, i int) string {
		return fmt.Sprintf("%s=%5.1f%%", move, probs[i]*100)
	}), ", "))

	chosenMove := forwardMoveStrs[lib.SampleFromWeights(probs)]

	return client.MoveResponse{
		Move:  chosenMove,
		Shout: "I'm moving " + chosenMove,
	}
}

func (sa *SnakeAgent) weightedScoresForHeuristic(heuristic WeightedHeuristic, nextStatesMap map[string][]GameSnapshot, forwardMoveStrs []string) map[string]float64 {
	moveScores := make(map[string]float64)
	for move, states := range nextStatesMap {
		moveScores[move] = lo.MeanBy(states, heuristic.F())
	}

	log.Printf("MoveScores for %25s: %s", heuristic.NameAndWeight(), strings.Join(lo.Map(forwardMoveStrs, func(move string, _ int) string {
		return fmt.Sprintf("%s=%6.1f", move, moveScores[move])
	}), ", "))

	weightedScores := lo.MapValues(moveScores, func(score float64, _ string) float64 {
		return score * heuristic.Weight()
	})

	return weightedScores
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
