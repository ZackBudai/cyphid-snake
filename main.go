package main

// Welcome to
// __________         __    __  .__                               __
// \______   \_____ _/  |__/  |_|  |   ____   ______ ____ _____  |  | __ ____
//  |    |  _/\__  \\   __\   __\  | _/ __ \ /  ___//    \\__  \ |  |/ // __ \
//  |    |   \ / __ \|  |  |  | |  |_\  ___/ \___ \|   |  \/ __ \|    <\  ___/
//  |________/(______/__|  |__| |____/\_____>______>___|__(______/__|__\\_____>
//
// This file can be a nice home for your Battlesnake logic and helper functions.
//
// To get you started we've included code to prevent your Battlesnake from moving backwards.
// For more info see https://docs.battlesnake.com/

import (
	"log"
	"math"
	"math/rand"
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

// info is called when you create your Battlesnake on play.battlesnake.com
// and controls your Battlesnake's appearance
// TIP: If you open your Battlesnake URL in a browser you should see this data
func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "crazycat911", // TODO: Your Battlesnake username
		Color:      "#f1f1f1",     // TODO: Choose color
		Head:       "snow-worm",     // TODO: Choose head
		Tail:       "block-bum",     // TODO: Choose tail
	}
}

// start is called when your Battlesnake begins a game
func start(state GameState) {
	log.Println("GAME START")
}

// end is called when your Battlesnake finishes a game
func end(state GameState) {
	log.Printf("GAME OVER\n\n")
}

// move is called on every turn and returns your next move
// Valid moves are "up", "down", "left", or "right"
// See https://docs.battlesnake.com/api/example-move for available data
func move(state GameState) BattlesnakeMoveResponse {

	isMoveSafe := map[string]bool{
		"up":    true,
		"down":  true,
		"left":  true,
		"right": true,
	}

	// We've included code to prevent your Battlesnake from moving backwards
	myHead := state.You.Body[0] // Coordinates of your head
	myNeck := state.You.Body[1] // Coordinates of your "neck"

	if myNeck.X < myHead.X { // Neck is left of head, don't move left
		isMoveSafe["left"] = false

	} else if myNeck.X > myHead.X { // Neck is right of head, don't move right
		isMoveSafe["right"] = false

	} else if myNeck.Y < myHead.Y { // Neck is below head, don't move down
		isMoveSafe["down"] = false

	} else if myNeck.Y > myHead.Y { // Neck is above head, don't move up
		isMoveSafe["up"] = false
	}

	// TODO: Step 1 - Prevent your Battlesnake from moving out of bounds
	boardWidth := state.Board.Width
	boardHeight := state.Board.Height

	if myHead.X == (boardWidth - 1) {
		isMoveSafe["right"] = false
	} else if myHead.X == 0 {
		isMoveSafe["left"] = false
	}

	if myHead.Y == (boardHeight - 1) {
		isMoveSafe["up"] = false
	} else if myHead.Y == 0 {
		isMoveSafe["down"] = false
	}

	// TODO: Step 2 & 3 - Prevent your Battlesnake from colliding with other Battlesnakes (and itself)
	snakes := state.Board.Snakes

	for _, snake := range snakes {
		for _, bodyPart := range snake.Body {
			if bodyPart.X == myHead.X && bodyPart.Y == myHead.Y {
				continue
			}
			if bodyPart.X == myHead.X-1 && bodyPart.Y == myHead.Y {
				isMoveSafe["left"] = false
			} else if bodyPart.X == myHead.X+1 && bodyPart.Y == myHead.Y {
				isMoveSafe["right"] = false
			}

			if bodyPart.X == myHead.X && bodyPart.Y == myHead.Y-1 {
				isMoveSafe["down"] = false
			} else if bodyPart.X == myHead.X && bodyPart.Y == myHead.Y+1 {
				isMoveSafe["up"] = false
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

	// Choose a random move from the safe ones
	nextMove := safeMoves[rand.Intn(len(safeMoves))]

	// TODO: Step 4 - Move towards food instead of random, to regain health and survive longer
	food := state.Board.Food

	closestFood := FindClosestFood(myHead, food)

	dx := closestFood.X - myHead.X
	dy := closestFood.Y - myHead.Y

	var goodmove string

	if dx > 0 {
		goodmove = "right"
	} else if dx < 0 {
		goodmove = "left"
	} else if dy > 0 {
		goodmove = "up"
	} else if dy < 0 {
		goodmove = "down"
	}

	if isMoveSafe[goodmove] {
		log.Printf("MOVE %d: %s\n", state.Turn, goodmove)
		return BattlesnakeMoveResponse{Move: goodmove}
	}

	log.Printf("MOVE %d: %s\n", state.Turn, nextMove)
	return BattlesnakeMoveResponse{Move: nextMove}
}

func main() {
	RunServer()
}
