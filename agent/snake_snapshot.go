package agent

import (
	"github.com/BattlesnakeOfficial/rules"
	// "github.com/samber/mo"
	"github.com/samber/lo"
	// "log"
)

type SnakeSnapshot interface {
	ID() string
	Name() string
	Health() int
	Body() []rules.Point
	Head() rules.Point
	Length() int
	LastShout() string
	ForwardMoves() []rules.SnakeMove
}

// SnakeSnapshot interface implementation
type snakeStatsImpl struct {
	name            string
	lastShout       string
	turnLastShouted int
}

type snakeSnapshotImpl struct {
	stats *snakeStatsImpl
	snake *rules.Snake
}

func (s *snakeSnapshotImpl) ID() string {
	return s.snake.ID
}

func (s *snakeSnapshotImpl) Name() string {
	return s.stats.name
}

func (s *snakeSnapshotImpl) Health() int {
	if s.snake.EliminatedCause != rules.NotEliminated {
		return 0
	} else {
		return s.snake.Health
	}
}

func (s *snakeSnapshotImpl) Body() []rules.Point {
	return s.snake.Body
}

func (s *snakeSnapshotImpl) Head() rules.Point {
	return s.snake.Body[0]
}

func (s *snakeSnapshotImpl) Length() int {
	return len(s.snake.Body)
}

func (s *snakeSnapshotImpl) LastShout() string {
	return s.stats.lastShout
}

func (s *snakeSnapshotImpl) ForwardMoves() []rules.SnakeMove {
	possibleMoveStrs := []string{"up", "down", "left", "right"}

	isBackwardMove := func(move string) bool {
		head := s.Head()
		neck := s.Body()[1]
		switch {
		case head.X < neck.X && move == "right":
			return true
		case head.X > neck.X && move == "left":
			return true
		case head.Y < neck.Y && move == "up":
			return true
		case head.Y > neck.Y && move == "down":
			return true
		}
		return false
	}

	if s.Length() == 1 {
		return lo.Map(possibleMoveStrs, func(move string, _ int) rules.SnakeMove {
			return rules.SnakeMove{ID: s.ID(), Move: move}
		})
	}

	forwardMoves := lo.FilterMap(possibleMoveStrs, func(move string, _ int) (rules.SnakeMove, bool) {
		return rules.SnakeMove{ID: s.ID(), Move: move}, !isBackwardMove(move)
	})

	return forwardMoves
}
