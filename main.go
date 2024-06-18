package main

import (
	"log"
	"math"
)


func ManhattanDistance(p1, p2 Coord) int {
	return int(math.Abs(float64(p1.X-p2.X)) + math.Abs(float64(p1.Y-p2.Y)))
}

func FindClosestFood(start Coord, foodPoints []Coord) Coord {
	if len(foodPoints) == 0 {
		return Coord{} // Returning a default Point if the foodPoints list is empty
	}

	closest := foodPoints[0]
	minDistance := ManhattanDistance(start, closest)

	for _, food := range foodPoints[1:] {
		distance := ManhattanDistance(start, food)
		if distance < minDistance {
			closest = food
			minDistance = distance
		}
	}

	return closest
}

func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "crazycat911", // TODO: Your Battlesnake username
		Color:      "#07e3a5",     // TODO: Choose color
		Head:       "gamer",       // TODO: Choose head
		Tail:       "coffee",      // TODO: Choose tail
	}
}

func start(state GameState) {
	log.Println("GAME START")
}

func end(state GameState) {
	log.Printf("GAME OVER\n\n")
}

func detect_danger(state GameState) []Coord {
	mySnake := state.You
	snakes := state.Board.Snakes
	danger_zones := []Coord{}

	for _, snake := range snakes {
		for _, bodypart := range snake.Body {
			if bodypart == snake.Body[snake.Length - 1] {
				if snake.Health == 100 {
					danger_zones = append(danger_zones, bodypart)
				}
			} else {
				danger_zones = append(danger_zones, bodypart)
			}
		}

		if !(snake.ID == mySnake.ID) {
			if snake.Length >= mySnake.Length {
				danger_zones = append(danger_zones, Coord{X: snake.Head.X + 1, Y: snake.Head.Y})
				danger_zones = append(danger_zones, Coord{X: snake.Head.X - 1, Y: snake.Head.Y})
				danger_zones = append(danger_zones, Coord{X: snake.Head.X, Y: snake.Head.Y + 1})
				danger_zones = append(danger_zones, Coord{X: snake.Head.X, Y: snake.Head.Y - 1})
			}
		}
	}
	return danger_zones
}

func isCoordInBoard(c Coord, board Board) bool {
	return c.X >= 0 && c.Y >= 0 && c.X < board.Width && c.Y < board.Height
}

func floodFill(start Coord, board Board, dangerZones map[Coord]bool) int {
	queue := []Coord{start}
	visited := map[Coord]bool{start: true}
	area := 0
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		area++

		directions := []Coord{
			{X: current.X + 1, Y: current.Y},
			{X: current.X - 1, Y: current.Y},
			{X: current.X, Y: current.Y + 1},
			{X: current.X, Y: current.Y - 1},
		}

		for _, neighbor := range directions {
			if isCoordInBoard(neighbor, board) && !visited[neighbor] && !dangerZones[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	return area
}

func move(state GameState) BattlesnakeMoveResponse {
	myHead := state.You.Body[0] // Coordinates of your head

	isMoveSafe := map[string]bool{
		"up":    true,
		"down":  true,
		"left":  true,
		"right": true,
	}

	if myHead.X == 0 {
		isMoveSafe["left"] = false
	} else if myHead.X == state.Board.Width-1 {
		isMoveSafe["right"] = false
	}

	if myHead.Y == 0 {
		isMoveSafe["down"] = false
	} else if myHead.Y == state.Board.Height-1 {
		isMoveSafe["up"] = false
	}

	dangerZones := detect_danger(state)
	dangerMap := make(map[Coord]bool)
	for _, dangerZone := range dangerZones {
		dangerMap[dangerZone] = true
	}

	for _, dangerZone := range dangerZones {
		if (Coord{X: myHead.X + 1, Y: myHead.Y}) == dangerZone {
			isMoveSafe["right"] = false
		} else if (Coord{X: myHead.X - 1, Y: myHead.Y}) == dangerZone {
			isMoveSafe["left"] = false
		} else if (Coord{X: myHead.X, Y: myHead.Y + 1}) == dangerZone {
			isMoveSafe["up"] = false
		} else if (Coord{X: myHead.X, Y: myHead.Y - 1}) == dangerZone {
			isMoveSafe["down"] = false
		}
	}

	// Ensure moves don't lead to an area smaller than the snake's length
	for move, isSafe := range isMoveSafe {
		if isSafe {
			var newHead Coord
			switch move {
			case "up":
				newHead = Coord{X: myHead.X, Y: myHead.Y + 1}
			case "down":
				newHead = Coord{X: myHead.X, Y: myHead.Y - 1}
			case "left":
				newHead = Coord{X: myHead.X - 1, Y: myHead.Y}
			case "right":
				newHead = Coord{X: myHead.X + 1, Y: myHead.Y}
			}

			area := floodFill(newHead, state.Board, dangerMap)
			if area < state.You.Length {
				isMoveSafe[move] = false
			}
		}
	}

	// Are there any safe moves left?
	safeMoves := []string{}
	for move, isSafe := range isMoveSafe {
		if isSafe {
			safeMoves = append(safeMoves, move)
		}
	}

	if len(safeMoves) == 0 {
		log.Printf("MOVE %d: No safe moves detected :( Moving up\n", state.Turn)
		return BattlesnakeMoveResponse{Move: "up"}
	}

	// Follow the closest food
	food := state.Board.Food
	if len(food) > 0 {
		closestFood := FindClosestFood(myHead, food)
		dx := closestFood.X - myHead.X
		dy := closestFood.Y - myHead.Y

		var moveTowardsFood string

		if dx > 0 && isMoveSafe["right"] {
			moveTowardsFood = "right"
		} else if dx < 0 && isMoveSafe["left"] {
			moveTowardsFood = "left"
		} else if dy > 0 && isMoveSafe["up"] {
			moveTowardsFood = "up"
		} else if dy < 0 && isMoveSafe["down"] {
			moveTowardsFood = "down"
		}

		if moveTowardsFood != "" {
			log.Printf("MOVE %d: Moving towards food %s\n", state.Turn, moveTowardsFood)
			return BattlesnakeMoveResponse{Move: moveTowardsFood}
		}
	}

	// If no food following move is possible, pick the safest move
	bestMove := safeMoves[0]
	maxArea := -1

	for _, move := range safeMoves {
		var newHead Coord
		switch move {
		case "up":
			newHead = Coord{X: myHead.X, Y: myHead.Y + 1}
		case "down":
			newHead = Coord{X: myHead.X, Y: myHead.Y - 1}
		case "left":
			newHead = Coord{X: myHead.X - 1, Y: myHead.Y}
		case "right":
			newHead = Coord{X: myHead.X + 1, Y: myHead.Y}
		}

		area := floodFill(newHead, state.Board, dangerMap)
		if area > maxArea {
			maxArea = area
			bestMove = move
		}
	}

	log.Printf("MOVE %d: %s\n", state.Turn, bestMove)
	return BattlesnakeMoveResponse{Move: bestMove}
}

func main() {
	RunServer()
}