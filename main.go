package main

import (
	"log"
	"container/heap"
	"math"
)

func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "zuthan",
		Color:      "#8B0000",
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

// Define directions
var directions = []string{"up", "down", "left", "right"}

// PriorityQueue for A* algorithm
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Node)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// Node for pathfinding algorithms
type Node struct {
	coord     Coord
	parent    *Node
	g         int
	h         int
	f         int
	priority  int
	direction string
}

func move(state GameState) BattlesnakeMoveResponse {
	if state.You.Health < 50 {
		return moveTowardsFood(state)
	}
	return moveTowardsTail(state)
}

func moveTowardsFood(state GameState) BattlesnakeMoveResponse {
	start := state.You.Head
	var closestFood Coord
	minDist := math.MaxInt32
	var nextMove string

	for _, food := range state.Board.Food {
		path, dist := bfs(state, start, food)
		if dist < minDist {
			minDist = dist
			closestFood = food
			if len(path) > 0 {
				nextMove = path[0]
			}
		}
	}

	if nextMove == "" {
		// If no path to food is found, move towards the closest food using Manhattan distance
		for _, food := range state.Board.Food {
			dist := manhattanDistance(start, food)
			if dist < minDist {
				minDist = dist
				closestFood = food
			}
		}
		nextMove = moveTowards(start, closestFood)
	}

	if nextMove == "" {
		nextMove = safestMove(state)
	}

	return BattlesnakeMoveResponse{
		Move:  nextMove,
		Shout: "Moving towards food!",
	}
}

func moveTowards(from, to Coord) string {
	if from.X < to.X {
		return "right"
	} else if from.X > to.X {
		return "left"
	} else if from.Y < to.Y {
		return "up"
	} else if from.Y > to.Y {
		return "down"
	}
	return ""
}



func moveTowardsTail(state GameState) BattlesnakeMoveResponse {
	start := state.You.Head
	tail := state.You.Body[len(state.You.Body)-1]

	path, _ := aStar(state, start, tail)
	if len(path) > 0 {
		return BattlesnakeMoveResponse{
			Move:  path[0],
			Shout: "Moving towards my tail!",
		}
	}

	// If no path to tail, move to the largest open space
	largestRegion := findLargestRegion(state)
	if len(largestRegion) > 0 {
		move := moveWithinRegion(state, largestRegion)
		return BattlesnakeMoveResponse{
			Move:  move,
			Shout: "Moving to open space!",
		}
	}

	// If all else fails, make a safe move
	return BattlesnakeMoveResponse{
		Move:  safestMove(state),
		Shout: "Making a safe move!",
	}
}

func bfs(state GameState, start, goal Coord) ([]string, int) {
	queue := []Node{{coord: start}}
	visited := make(map[Coord]bool)
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.coord == goal {
			return reconstructPath(current), len(reconstructPath(current))
		}

		for _, dir := range directions {
			next := moveCoord(current.coord, dir)
			if !visited[next] && isValidMove(state, next) {
				visited[next] = true
				newNode := Node{coord: next, parent: &current, direction: dir}
				queue = append(queue, newNode)
			}
		}
	}

	return nil, math.MaxInt32
}

func aStar(state GameState, start, goal Coord) ([]string, int) {
	openSet := &PriorityQueue{}
	heap.Init(openSet)
	startNode := &Node{coord: start, g: 0, h: manhattanDistance(start, goal), f: 0}
	heap.Push(openSet, startNode)

	cameFrom := make(map[Coord]*Node)
	gScore := make(map[Coord]int)
	gScore[start] = 0

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)

		if current.coord == goal {
			return reconstructPath(*current), current.g
		}

		for _, dir := range directions {
			neighbor := moveCoord(current.coord, dir)
			if !isValidMove(state, neighbor) {
				continue
			}

			tentativeGScore := gScore[current.coord] + getEdgeWeight(state, current.coord, neighbor)

			if gScore[neighbor] == 0 || tentativeGScore < gScore[neighbor] {
				cameFrom[neighbor] = current
				gScore[neighbor] = tentativeGScore
				fScore := tentativeGScore + manhattanDistance(neighbor, goal)
				neighborNode := &Node{
					coord:     neighbor,
					parent:    current,
					g:         tentativeGScore,
					h:         manhattanDistance(neighbor, goal),
					f:         fScore,
					priority:  fScore,
					direction: dir,
				}
				heap.Push(openSet, neighborNode)
			}
		}
	}

	return nil, math.MaxInt32
}

func findLargestRegion(state GameState) map[Coord]bool {
	largestRegion := make(map[Coord]bool)
	visited := make(map[Coord]bool)

	for y := 0; y < state.Board.Height; y++ {
		for x := 0; x < state.Board.Width; x++ {
			coord := Coord{X: x, Y: y}
			if !visited[coord] && isValidMove(state, coord) {
				region := floodFill(state, coord, visited)
				if len(region) > len(largestRegion) {
					largestRegion = region
				}
			}
		}
	}

	return largestRegion
}

func floodFill(state GameState, start Coord, visited map[Coord]bool) map[Coord]bool {
	region := make(map[Coord]bool)
	queue := []Coord{start}
	visited[start] = true
	region[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dir := range directions {
			next := moveCoord(current, dir)
			if !visited[next] && isValidMove(state, next) {
				visited[next] = true
				region[next] = true
				queue = append(queue, next)
			}
		}
	}

	return region
}

func moveWithinRegion(state GameState, region map[Coord]bool) string {
	head := state.You.Head
	maxSpace := 0
	bestMove := ""

	for _, dir := range directions {
		next := moveCoord(head, dir)
		if region[next] {
			space := countAccessibleSpace(state, next, region)
			if space > maxSpace {
				maxSpace = space
				bestMove = dir
			}
		}
	}

	if bestMove == "" {
		return safestMove(state)
	}

	return bestMove
}

func countAccessibleSpace(state GameState, start Coord, region map[Coord]bool) int {
	visited := make(map[Coord]bool)
	queue := []Coord{start}
	visited[start] = true
	count := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		count++

		for _, dir := range directions {
			next := moveCoord(current, dir)
			if !visited[next] && region[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}

	return count
}

func safestMove(state GameState) string {
	head := state.You.Head
	safeMoves := make(map[string]bool)

	for _, dir := range directions {
		next := moveCoord(head, dir)
		if isValidMove(state, next) {
			safeMoves[dir] = true
		}
	}

	if len(safeMoves) == 0 {
		return "up" // No safe moves, just go up
	}

	// Prefer moves that don't reduce options
	for _, dir := range directions {
		if safeMoves[dir] {
			next := moveCoord(head, dir)
			if countSafeMoves(state, next) >= len(safeMoves) {
				return dir
			}
		}
	}

	// If all moves reduce options, pick the first safe move
	for _, dir := range directions {
		if safeMoves[dir] {
			return dir
		}
	}

	return "up" // Fallback
}

func countSafeMoves(state GameState, coord Coord) int {
	count := 0
	for _, dir := range directions {
		next := moveCoord(coord, dir)
		if isValidMove(state, next) {
			count++
		}
	}
	return count
}

func isValidMove(state GameState, coord Coord) bool {
	// Check if the coordinate is within the board
	if coord.X < 0 || coord.X >= state.Board.Width || coord.Y < 0 || coord.Y >= state.Board.Height {
		return false
	}

	// Check if the coordinate is occupied by a snake
	for _, snake := range state.Board.Snakes {
		for _, body := range snake.Body {
			if body == coord {
				return false
			}
		}
	}

	return true
}

func getEdgeWeight(state GameState, from, to Coord) int {
	for _, snake := range state.Board.Snakes {
		if snake.ID != state.You.ID && manhattanDistance(snake.Head, to) == 1 && len(snake.Body) >= len(state.You.Body) {
			return 4 // Risky move, potential head-to-head collision
		}
	}
	return 1 // Safe move
}

func moveCoord(coord Coord, direction string) Coord {
	switch direction {
	case "up":
		return Coord{X: coord.X, Y: coord.Y + 1}
	case "down":
		return Coord{X: coord.X, Y: coord.Y - 1}
	case "left":
		return Coord{X: coord.X - 1, Y: coord.Y}
	case "right":
		return Coord{X: coord.X + 1, Y: coord.Y}
	}
	return coord
}

func manhattanDistance(a, b Coord) int {
	return int(math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y)))
}

func reconstructPath(node Node) []string {
	path := []string{}
	current := &node
	for current.parent != nil {
		path = append([]string{current.direction}, path...)
		current = current.parent
	}
	return path
}

func main() {
	RunServer()
}
