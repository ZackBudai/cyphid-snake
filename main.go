package main

import (
	"log"
	"math"
	"sort"
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
// Directions
var directions = map[string]Coord{
	"up":    {X: 0, Y: 1},
	"down":  {X: 0, Y: -1},
	"left":  {X: -1, Y: 0},
	"right": {X: 1, Y: 0},
}

func move(state GameState) BattlesnakeMoveResponse {
	possibleMoves := getPossibleMoves(state)
	if len(possibleMoves) == 0 {
		return BattlesnakeMoveResponse{Move: "up", Shout: "No safe moves!"}
	}

	move := getBestMove(state, possibleMoves)
	return BattlesnakeMoveResponse{Move: move, Shout: "Moving " + move + "!"}
}

func getPossibleMoves(state GameState) []string {
	possibleMoves := []string{}
	myHead := state.You.Head

	for direction, offset := range directions {
		newPos := Coord{X: myHead.X + offset.X, Y: myHead.Y + offset.Y}
		if isSafeMove(newPos, state) {
			possibleMoves = append(possibleMoves, direction)
		}
	}

	return possibleMoves
}

func isSafeMove(pos Coord, state GameState) bool {
	// Check board boundaries
	if pos.X < 0 || pos.Y < 0 || pos.X >= state.Board.Width || pos.Y >= state.Board.Height {
		return false
	}

	// Check collision with own body
	for _, bodyPart := range state.You.Body[1:] {
		if pos.X == bodyPart.X && pos.Y == bodyPart.Y {
			return false
		}
	}

	// Check collision with other snakes
	for _, snake := range state.Board.Snakes {
		for _, bodyPart := range snake.Body {
			if pos.X == bodyPart.X && pos.Y == bodyPart.Y {
				return false
			}
		}
	}

	return true
}

func getBestMove(state GameState, possibleMoves []string) string {
	type moveScore struct {
		move  string
		score float64
	}

	var scoredMoves []moveScore

	for _, move := range possibleMoves {
		score := evaluateMove(state, move)
		scoredMoves = append(scoredMoves, moveScore{move, score})
	}

	sort.Slice(scoredMoves, func(i, j int) bool {
		return scoredMoves[i].score > scoredMoves[j].score
	})

	return scoredMoves[0].move
}

func evaluateMove(state GameState, move string) float64 {
	newHead := Coord{
		X: state.You.Head.X + directions[move].X,
		Y: state.You.Head.Y + directions[move].Y,
	}

	score := 0.0

	// Survival score
	score += 10.0

	// Food seeking score (now using a more nuanced health factor)
	foodScore := evaluateFoodScore(state, newHead)
	healthFactor := evaluateHealthManagement(state)
	score += foodScore * healthFactor

	// Space control score
	score += evaluateSpaceControl(state, newHead)

	// Squad awareness score
	score += evaluateSquadAwareness(state, newHead)

	// Opponent avoidance score
	score += evaluateOpponentAvoidance(state, newHead)

	return score
}

func evaluateFoodScore(state GameState, newHead Coord) float64 {
	totalScore := 0.0
	maxDistance := float64(state.Board.Width + state.Board.Height) // Max possible distance on the board

	for _, food := range state.Board.Food {
		distance := pathDistance(state, newHead, food)
		if distance == -1 {
			continue // Skip unreachable food
		}

		// Score decreases as distance increases, but never reaches zero
		// This ensures that distant food still contributes, but much less than nearby food
		score := maxDistance / (float64(distance) + 1)
		totalScore += score
	}

	// Normalize the score based on the board size and number of food items
	normalizedScore := totalScore / (float64(len(state.Board.Food)) * maxDistance)

	// Scale the normalized score to a reasonable range (e.g., 0 to 10)
	return normalizedScore * 10
}

func evaluateSpaceControl(state GameState, newHead Coord) float64 {
	floodFillResult := floodFill(state, newHead)
	return float64(floodFillResult) / float64(state.Board.Width*state.Board.Height) * 10
}

func evaluateSquadAwareness(state GameState, newHead Coord) float64 {
	score := 0.0
	for _, snake := range state.Board.Snakes {
		if snake.ID != state.You.ID && isTeammate(state.You, snake) {
			distance := manhattanDistance(newHead, snake.Head)
			if distance < 2 {
				score -= 10.0 // Avoid getting too close to teammates
			}
		}
	}
	return score
}

func evaluateHealthManagement(state GameState) float64 {
	health := float64(state.You.Health)

	// Base factor: ranges from 1.0 (at full health) to 2.0 (at 0 health)
	baseFactor := 1.0 + (100.0 - health) / 100.0

	// Urgency factor: increases rapidly at very low health
	urgencyThreshold := 25.0
	urgencyFactor := 1.0
	if health < urgencyThreshold {
		urgencyFactor = 1.0 + math.Pow((urgencyThreshold-health)/urgencyThreshold, 2)
	}

	// Combine base factor and urgency factor
	return baseFactor * urgencyFactor
}

func evaluateOpponentAvoidance(state GameState, newHead Coord) float64 {
	score := 0.0
	for _, snake := range state.Board.Snakes {
		if snake.ID != state.You.ID && !isTeammate(state.You, snake) {
			distance := manhattanDistance(newHead, snake.Head)
			if distance == 0 {
				score -= 100 // Heavily penalize head-on collisions
			} else if distance == 1 {
				if len(state.You.Body) <= len(snake.Body) {
					score -= 50 // Penalize risky head-to-head standoffs
				} else {
					score += 10 // Reward potential eliminations
				}
			}
		}
	}
	return score
}

func manhattanDistance(a, b Coord) float64 {
	return math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y))
}

func pathDistance(state GameState, start, end Coord) int {
	visited := make(map[Coord]bool)
	queue := []struct {
		pos   Coord
		steps int
	}{{pos: start, steps: 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.pos == end {
			return current.steps
		}

		if visited[current.pos] {
			continue
		}
		visited[current.pos] = true

		for _, dir := range directions {
			next := Coord{X: current.pos.X + dir.X, Y: current.pos.Y + dir.Y}
			if isSafeMove(next, state) && !visited[next] {
				queue = append(queue, struct {
					pos   Coord
					steps int
				}{pos: next, steps: current.steps + 1})
			}
		}
	}

	return -1 // No path found
}

func floodFill(state GameState, start Coord) int {
	visited := make(map[Coord]bool)
	stack := []Coord{start}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}

		visited[current] = true

		for _, dir := range directions {
			next := Coord{X: current.X + dir.X, Y: current.Y + dir.Y}
			if isSafeMove(next, state) && !visited[next] {
				stack = append(stack, next)
			}
		}
	}

	return len(visited)
}

func isTeammate(snake1, snake2 Battlesnake) bool {
	// In a real implementation, you'd need a way to identify teammates.
	// For this example, we'll consider snakes with the same color as teammates.
	return snake1.Customizations.Color == snake2.Customizations.Color
}
