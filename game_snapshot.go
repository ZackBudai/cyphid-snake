package main

import (
	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/client"

	// "encoding/json"
	"log"
)

type GameSnapshot interface {
	GameID() string
	Turn() int
	Height() int
	Width() int
	Food() []client.Coord
	Hazards() []client.Coord
	Snakes() []SnakeSnapshot
	You() SnakeSnapshot
	Rules() client.Ruleset
	Teammates() []SnakeSnapshot
	YourTeam() []SnakeSnapshot
	Opponents() []SnakeSnapshot
	ApplyMoves(moves []rules.SnakeMove) (GameSnapshot, error)
}

type SnakeSnapshot interface {
	ID() string
	Name() string
	Health() int
	Body() []client.Coord
	Head() client.Coord
	Length() int
	Shout() string
	Squad() string
	Latency() string
	Color() string
	ForwardMoves() []rules.SnakeMove
}

type gameSnapshotImpl struct {
	request    *client.SnakeRequest
	boardState *rules.BoardState
	ruleset    rules.Ruleset
}

type snakeSnapshotImpl struct {
	originalSnake *client.Snake
	newSnake      *rules.Snake
}

// GameSnapshot interface implementation

func (g *gameSnapshotImpl) GameID() string {
	return g.request.Game.ID
}

func (g *gameSnapshotImpl) Turn() int {
	if g.boardState != nil {
		return g.boardState.Turn
	}
	return g.request.Turn
}

func (g *gameSnapshotImpl) Height() int {
	if g.boardState != nil {
		return g.boardState.Height
	}
	return g.request.Board.Height
}

func (g *gameSnapshotImpl) Width() int {
	if g.boardState != nil {
		return g.boardState.Width
	}
	return g.request.Board.Width
}

func (g *gameSnapshotImpl) Food() []client.Coord {
	if g.boardState != nil {
		return pointsToCoords(g.boardState.Food)
	}
	return g.request.Board.Food
}

func (g *gameSnapshotImpl) Hazards() []client.Coord {
	if g.boardState != nil {
		return pointsToCoords(g.boardState.Hazards)
	}
	return g.request.Board.Hazards
}

func (g *gameSnapshotImpl) Snakes() []SnakeSnapshot {
	var snakes []SnakeSnapshot
	if g.boardState != nil {
		for _, newSnake := range g.boardState.Snakes {
			originalSnake := g.findSnakeByID(newSnake.ID)
			snakes = append(snakes, &snakeSnapshotImpl{originalSnake: originalSnake, newSnake: &newSnake})
		}
	} else {
		for _, snake := range g.request.Board.Snakes {
			snakes = append(snakes, &snakeSnapshotImpl{originalSnake: &snake, newSnake: nil})
		}
	}
	return snakes
}

func (g *gameSnapshotImpl) You() SnakeSnapshot {
	youID := g.request.You.ID
	for _, snake := range g.Snakes() {
		if snake.ID() == youID {
			return snake
		}
	}
	return nil
}

func (g *gameSnapshotImpl) Rules() client.Ruleset {
	return g.request.Game.Ruleset
}

func (g *gameSnapshotImpl) findSnakeByID(id string) *client.Snake {
	for _, snake := range g.request.Board.Snakes {
		if snake.ID == id {
			return &snake
		}
	}
	return nil
}

func (g *gameSnapshotImpl) Teammates() []SnakeSnapshot {
	you := g.You()
	var teammates []SnakeSnapshot

	for _, snake := range g.Snakes() {
		if snake.ID() != you.ID() {
			if (you.Squad() != "" && snake.Squad() == you.Squad()) ||
				(you.Squad() == "" && snake.Color() == you.Color()) {
				teammates = append(teammates, snake)
			}
		}
	}

	return teammates
}

func (g *gameSnapshotImpl) YourTeam() []SnakeSnapshot {
	team := []SnakeSnapshot{g.You()}
	team = append(team, g.Teammates()...)
	return team
}

func (g *gameSnapshotImpl) Opponents() []SnakeSnapshot {
	you := g.You()
	var opponents []SnakeSnapshot

	for _, snake := range g.Snakes() {
		if snake.ID() != you.ID() {
			if (you.Squad() != "" && snake.Squad() != you.Squad()) ||
				(you.Squad() == "" && snake.Color() != you.Color()) {
				opponents = append(opponents, snake)
			}
		}
	}

	return opponents
}

func (g *gameSnapshotImpl) ApplyMoves(moves []rules.SnakeMove) (GameSnapshot, error) {
	if len(moves) == 0 {
		log.Fatalf("No moves provided: %+v", moves)
	}

	if g.ruleset == nil {
		log.Fatalf("Ruleset is nil")
	}

	if g.boardState == nil {
		log.Fatalf("BoardState is nil")
	}

	_, nextBoardState, err := g.ruleset.Execute(g.boardState, moves)
	if err != nil {
		log.Printf("Error executing moves: %v", err)
		return nil, err
	}
	return &gameSnapshotImpl{
		request:    g.request,
		boardState: nextBoardState,
		ruleset:    g.ruleset,
	}, nil
}

func NewGameSnapshot(request *client.SnakeRequest) GameSnapshot {
	if request == nil {
		log.Println("Error: Request is nil")
		return nil
	}
	boardState := ConvertToBoardState(*request)

	builder := rules.NewRulesetBuilder()
	ruleset := builder.NamedRuleset(request.Game.Ruleset.Name)

	if ruleset == nil {
		log.Printf("Failed to create ruleset for request: %v", request.Game.Ruleset.Name)
		return nil
	}

	return &gameSnapshotImpl{
		request:    request,
		boardState: boardState,
		ruleset:    ruleset,
	}
}

// SnakeSnapshot interface implementation

func (s *snakeSnapshotImpl) ID() string {
	return s.originalSnake.ID
}

func (s *snakeSnapshotImpl) Name() string {
	return s.originalSnake.Name
}

func (s *snakeSnapshotImpl) Health() int {
	if s.newSnake != nil {
		if s.newSnake.EliminatedCause != "" {
			// log.Printf("EliminatedCause: %s", s.newSnake.EliminatedCause)
			return 0
		} else {
			return s.newSnake.Health
		}
	}
	return s.originalSnake.Health
}

func (s *snakeSnapshotImpl) Body() []client.Coord {
	if s.newSnake != nil {
		return pointsToCoords(s.newSnake.Body)
	}
	return s.originalSnake.Body
}

func (s *snakeSnapshotImpl) Head() client.Coord {
	if s.newSnake != nil {
		return pointToCoord(s.newSnake.Body[0])
	}
	return s.originalSnake.Head
}

func (s *snakeSnapshotImpl) Length() int {
	if s.newSnake != nil {
		return len(s.newSnake.Body)
	}
	return s.originalSnake.Length
}

func (s *snakeSnapshotImpl) Shout() string {
	return s.originalSnake.Shout
}

func (s *snakeSnapshotImpl) Squad() string {
	return s.originalSnake.Squad
}

func (s *snakeSnapshotImpl) Latency() string {
	return s.originalSnake.Latency
}

func (s *snakeSnapshotImpl) Color() string {
	return s.originalSnake.Customizations.Color
}

func (s *snakeSnapshotImpl) ForwardMoves() []rules.SnakeMove {
	possibleMoveStrs := []string{"up", "down", "left", "right"}
	var forwardMoveStrs []string
	var forwardMoves []rules.SnakeMove
	var allMoves []rules.SnakeMove

	for _, move := range possibleMoveStrs {
		allMoves = append(allMoves, rules.SnakeMove{ID: s.ID(), Move: move})
	}

	// If it's the first turn (snake length is 1), all moves are possible
	if s.Length() == 1 {
		return allMoves
	}

	// Get the current head and neck positions
	head := s.Head()
	neck := s.Body()[1]

	// Determine the backward direction
	backwardMove := ""
	if head == neck {
		return allMoves
	}
	// log.Printf("Head: %+v, Neck: %+v", head, neck)
	if head.X < neck.X {
		backwardMove = "right"
	} else if head.X > neck.X {
		backwardMove = "left"
	} else if head.Y < neck.Y {
		backwardMove = "up"
	} else if head.Y > neck.Y {
		backwardMove = "down"
	}

	// Add all moves except the backward move
	for _, move := range possibleMoveStrs {
		if move != backwardMove {
			forwardMoveStrs = append(forwardMoveStrs, move)
			forwardMoves = append(forwardMoves, rules.SnakeMove{ID: s.ID(), Move: move})
		}
	}

	log.Printf("BackwardMove: %s, ForwardMoves: %+v", backwardMove, forwardMoveStrs)
	return forwardMoves
}
