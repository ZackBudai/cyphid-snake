package main

import (
	"log"
	"math"

	"github.com/BattlesnakeOfficial/rules"
)

// Update the SnakeAgent structure to include SnakeMetadataResponse
type SnakeAgent struct {
  portfolio Portfolio
  metadata  SnakeMetadataResponse
}
func NewSnakeAgent(portfolio Portfolio, metadata SnakeMetadataResponse) *SnakeAgent {
  return &SnakeAgent{
    portfolio: portfolio,
    metadata:  metadata,
  }
}

func (sa *SnakeAgent) ChooseMove(snapshot GameSnapshot) MoveResponse {
	forwardMoves := snapshot.You().ForwardMoves()
	scores := make([]float64, len(forwardMoves))

	for i, move := range forwardMoves {
		nextStates := sa.generateNextStates(snapshot, move.Move)
		if len(nextStates) == 0 {
			scores[i] = math.Inf(-1)
			continue
		}

		for _, heuristic := range sa.portfolio {
			marginalScore := sa.calculateMarginalScore(heuristic.Heuristic, nextStates)
			
			// Debug: Print Turn() index and marginalScore
			log.Printf("Turn: %d, Marginal Score: %f", snapshot.Turn(), marginalScore)
			scores[i] += marginalScore * heuristic.Weight
		}
	}

	chosenMove := forwardMoves[SoftmaxSample(scores)]
	return MoveResponse{
		Move:  chosenMove.Move,
		Shout: "I'm moving " + chosenMove.Move,
	}
}

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

func (sa *SnakeAgent) generateNextStates(snapshot GameSnapshot, move string) []GameSnapshot {
  var nextStates []GameSnapshot
  yourID := snapshot.You().ID()
  // Debugging info
  // log.Println("Starting generateNextStates")

  // Generate all possible move combinations for other snakes
  moveCombinations := generateMoveCombinations(snapshot.Snakes(), yourID)
  
  log.Printf("Trying move %s, combinations: %v", move, getMoveComboList(moveCombinations))
  
  for _, combination := range moveCombinations {
    combination[yourID] = rules.SnakeMove{ID: yourID, Move: move}
    // Convert the combination map to a slice
    var moveSlice []rules.SnakeMove
    for _, m := range combination {
      moveSlice = append(moveSlice, m)
    }
    // log.Printf("Move slice: %+v", moveSlice)
    // Debugging: Check before ApplyMoves call
    if snapshot == nil {
      log.Println("Snapshot is nil before applying moves")
    }
    nextState, err := snapshot.ApplyMoves(moveSlice)

    // Debug the state after ApplyMoves call
    if err != nil {
      log.Printf("Error applying moves: %v", err)
    } else {
      // log.Printf("Next state after applying move: %+v", nextState)
    }
    if nextState != nil {
      nextStates = append(nextStates, nextState)
    }
  }
  // log.Printf("Generated next states: %+v", nextStates)
  return nextStates
}

func generateMoveCombinations(snakes []SnakeSnapshot, excludeID string) []map[string]rules.SnakeMove {
  
  var combinations []map[string]rules.SnakeMove
  // Helper function to recursively build the move combinations
  var buildCombinations func(int, map[string]rules.SnakeMove)
  buildCombinations = func(index int, currentCombination map[string]rules.SnakeMove) {
      if index == len(snakes) {
          combinationCopy := make(map[string]rules.SnakeMove)
          for k, v := range currentCombination {
              combinationCopy[k] = v
          }
          combinations = append(combinations, combinationCopy)
          return
      }
      snake := snakes[index]
      if snake.ID() == excludeID {
          buildCombinations(index+1, currentCombination)
          return
      }
      for _, move := range snake.ForwardMoves() {
          currentCombination[snake.ID()] = move
          buildCombinations(index+1, currentCombination)
      }
  }
  buildCombinations(0, make(map[string]rules.SnakeMove))
  return combinations
}

func (sa *SnakeAgent) calculateMarginalScore(heuristic HeuristicFunc, nextStates []GameSnapshot) float64 {
  var totalScore float64
  for _, state := range nextStates {
    totalScore += float64(heuristic(state))
  }
  expectedScore := totalScore / float64(len(nextStates))
  // Calculate mean expected score across all moves (assuming 3 non-backward moves)
  meanExpectedScore := expectedScore / 3
  return expectedScore - meanExpectedScore
}
