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

	// Check hazards
	for _, hazard := range state.Board.Hazards {
		if pos.X == hazard.X && pos.Y == hazard.Y {
			// We'll allow hazards, but consider them less desirable
			return true
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

	// Food seeking score
	score += evaluateFoodScore(state, newHead)

	// Space control score
	score += evaluateSpaceControl(state, newHead)

	// Squad awareness score
	score += evaluateSquadAwareness(state, newHead)

	// Health management score
	score += evaluateHealthManagement(state)

	// Opponent avoidance score
	score += evaluateOpponentAvoidance(state, newHead)

	return score
}

func evaluateFoodScore(state GameState, newHead Coord) float64 {
	if state.You.Health > 50 {
		return 0 // Don't prioritize food if health is high
	}

	minDistance := math.Inf(1)
	for _, food := range state.Board.Food {
		distance := manhattanDistance(newHead, food)
		if distance < minDistance {
			minDistance = distance
		}
	}

	return 5.0 / (minDistance + 1)
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
				score -= 5.0 // Avoid getting too close to teammates
			}
		}
	}
	return score
}

func evaluateHealthManagement(state GameState) float64 {
	if state.You.Health < 25 {
		return 5.0 // Increase priority of food when health is low
	}
	return 0
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