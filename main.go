package main

import (
	"log"
	"math"
)

func main() {
	RunServer()
}

func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "zuthan",
		Color:      "#FF7F7F",
		Head:       "evil",
		Tail:       "nr-booster",
	}
}

func start(state GameState) {
	log.Println("GAME START")
}

func end(state GameState) {
	log.Printf("GAME OVER\n\n")
}

// ============ Implementation ==========

// Possible moves
var moves = []string{"up", "down", "left", "right"}

// move decides the next move for the Battlesnake
func move(state GameState) BattlesnakeMoveResponse {
	possibleMoves := getPossibleMoves(state)
	if len(possibleMoves) == 0 {
		return BattlesnakeMoveResponse{Move: "up", Shout: "No safe moves!"}
	}

	// Score moves based on various factors
	moveScores := make(map[string]int)
	for _, move := range possibleMoves {
		moveScores[move] = 0
		moveScores[move] += scoreFood(state, move)
		moveScores[move] += scoreSpace(state, move)
		moveScores[move] += scoreAggression(state, move)
		moveScores[move] -= scoreHazards(state, move)
	}

	// Choose the move with the highest score
	bestMove := possibleMoves[0]
	bestScore := moveScores[bestMove]
	for _, move := range possibleMoves[1:] {
		if moveScores[move] > bestScore {
			bestMove = move
			bestScore = moveScores[move]
		}
	}

	return BattlesnakeMoveResponse{Move: bestMove, Shout: "Moving " + bestMove}
}

// getPossibleMoves returns a list of safe moves
func getPossibleMoves(state GameState) []string {
	possibleMoves := []string{}
	head := state.You.Head
	for _, move := range moves {
		newPos := getNextPosition(head, move)
		if isSafe(newPos, state) {
			possibleMoves = append(possibleMoves, move)
		}
	}
	return possibleMoves
}

// isSafe checks if a position is safe to move to
func isSafe(pos Coord, state GameState) bool {
	// Check board boundaries
	if pos.X < 0 || pos.Y < 0 || pos.X >= state.Board.Width || pos.Y >= state.Board.Height {
		return false
	}

	// Check for collision with own body (except tail)
	for i, bodyPart := range state.You.Body[:len(state.You.Body)-1] {
		if i == 0 {
			continue // Skip head
		}
		if pos.X == bodyPart.X && pos.Y == bodyPart.Y {
			return false
		}
	}

	// Check for collision with other snakes
	for _, snake := range state.Board.Snakes {
		for _, bodyPart := range snake.Body {
			if pos.X == bodyPart.X && pos.Y == bodyPart.Y {
				return false
			}
		}
	}

	return true
}

// getNextPosition calculates the next position given a move
func getNextPosition(current Coord, move string) Coord {
	switch move {
	case "up":
		return Coord{X: current.X, Y: current.Y + 1}
	case "down":
		return Coord{X: current.X, Y: current.Y - 1}
	case "left":
		return Coord{X: current.X - 1, Y: current.Y}
	case "right":
		return Coord{X: current.X + 1, Y: current.Y}
	}
	return current
}

// scoreFood scores a move based on proximity to food
func scoreFood(state GameState, move string) int {
	nextPos := getNextPosition(state.You.Head, move)
	closestFoodDist := math.MaxInt32
	for _, food := range state.Board.Food {
		dist := manhattanDistance(nextPos, food)
		if dist < closestFoodDist {
			closestFoodDist = dist
		}
	}
	// Prioritize food more when health is low
	if state.You.Health < 25 {
		return 100 - closestFoodDist
	}
	return 50 - closestFoodDist
}

// scoreSpace scores a move based on the open space it leads to
func scoreSpace(state GameState, move string) int {
	nextPos := getNextPosition(state.You.Head, move)
	space := floodFill(nextPos, state)
	return space * 2
}

// scoreAggression scores a move based on aggressive potential
func scoreAggression(state GameState, move string) int {
	nextPos := getNextPosition(state.You.Head, move)
	score := 0
	for _, snake := range state.Board.Snakes {
		if snake.ID == state.You.ID {
			continue
		}
		if manhattanDistance(nextPos, snake.Head) == 1 && state.You.Length > snake.Length {
			score += 50
		}
	}
	return score
}

// scoreHazards scores a move based on proximity to hazards
func scoreHazards(state GameState, move string) int {
	nextPos := getNextPosition(state.You.Head, move)
	for _, hazard := range state.Board.Hazards {
		if nextPos.X == hazard.X && nextPos.Y == hazard.Y {
			return 30
		}
	}
	return 0
}

// manhattanDistance calculates the Manhattan distance between two coordinates
func manhattanDistance(a, b Coord) int {
	return int(math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y)))
}

// floodFill performs a flood fill to count accessible squares
func floodFill(start Coord, state GameState) int {
	visited := make(map[Coord]bool)
	queue := []Coord{start}
	count := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}

		visited[current] = true
		count++

		for _, move := range moves {
			next := getNextPosition(current, move)
			if isSafe(next, state) && !visited[next] {
				queue = append(queue, next)
			}
		}
	}

	return count
}
