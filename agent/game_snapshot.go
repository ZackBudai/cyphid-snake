package agent

import (
	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/client"
	"github.com/samber/lo"
	"github.com/samber/mo"
	// "encoding/json"
	"log"
)

type GameSnapshot interface {
	GameID() string
	Rules() rules.Ruleset
	Turn() int
	Height() int
	Width() int
	Food() []rules.Point
	Hazards() []rules.Point
	You() SnakeSnapshot
	Snakes() []SnakeSnapshot
	Teammates() []SnakeSnapshot
	YourTeam() []SnakeSnapshot
	Opponents() []SnakeSnapshot
	AllSnakes() []SnakeSnapshot
	DeadSnakes() []SnakeSnapshot
	ApplyMoves(moves []rules.SnakeMove) (GameSnapshot, error)
}

type gameSnapshotImpl struct {
	gameID      string
	ruleset     rules.Ruleset
	boardState  *rules.BoardState // must not be nil
	snakeStats  map[string]*snakeStatsImpl
	yourID      string
	allyIDs     []string
	opponentIDs []string
}

// GameSnapshot interface implementation

func (g *gameSnapshotImpl) GameID() string {
	return g.gameID
}

func (g *gameSnapshotImpl) Turn() int {
	return g.boardState.Turn
}

func (g *gameSnapshotImpl) Height() int {
	return g.boardState.Height
}

func (g *gameSnapshotImpl) Width() int {
	return g.boardState.Width
}

func (g *gameSnapshotImpl) Food() []rules.Point {
	return g.boardState.Food
}

func (g *gameSnapshotImpl) Hazards() []rules.Point {
	return g.boardState.Hazards
}

func (g *gameSnapshotImpl) AllSnakes() []SnakeSnapshot {
	return lo.Map(g.boardState.Snakes, func(snake rules.Snake, _ int) SnakeSnapshot {
		snakeStat := g.snakeStats[snake.ID]
		return &snakeSnapshotImpl{
			stats: snakeStat,
			snake: &snake,
		}
	})
}

func (g *gameSnapshotImpl) DeadSnakes() []SnakeSnapshot {
	return lo.Filter(g.AllSnakes(), func(s SnakeSnapshot, _ int) bool {
		return !s.Alive()
	})
}

func (g *gameSnapshotImpl) Snakes() []SnakeSnapshot {
	return lo.Filter(g.AllSnakes(), func(s SnakeSnapshot, _ int) bool {
		return s.Alive()
	})
}

func (g *gameSnapshotImpl) getSnakeById(id string) SnakeSnapshot {
	snake, found := lo.Find(g.boardState.Snakes, func(s rules.Snake) bool {
		return s.ID == id
	})
	if !found {
		panic("snake not found in boardState with id: " + id)
	}

	snakeStat, found := g.snakeStats[id]
	if !found {
		panic("snakeStats not found for id " + id)
	}

	return &snakeSnapshotImpl{
		stats: snakeStat,
		snake: &snake,
	}
}

func (g *gameSnapshotImpl) You() SnakeSnapshot {
	return g.getSnakeById(g.yourID)
}

func (g *gameSnapshotImpl) Rules() rules.Ruleset {
	return g.ruleset
}

func (g *gameSnapshotImpl) Teammates() []SnakeSnapshot {
	teammateIds := lo.Reject(g.allyIDs, func(id string, _ int) bool {
		return id == g.yourID
	})

	return lo.FilterMap(teammateIds, func(id string, _ int) (SnakeSnapshot, bool) {
		snakeSnapshot := g.getSnakeById(id)
		return snakeSnapshot, snakeSnapshot.Alive()
	})
}

func (g *gameSnapshotImpl) YourTeam() []SnakeSnapshot {
	return lo.FilterMap(g.allyIDs, func(id string, _ int) (SnakeSnapshot, bool) {
		snakeSnapshot := g.getSnakeById(id)
		return snakeSnapshot, snakeSnapshot.Alive()
	})
}

func (g *gameSnapshotImpl) Opponents() []SnakeSnapshot {
	return lo.FilterMap(g.opponentIDs, func(id string, _ int) (SnakeSnapshot, bool) {
		snakeSnapshot := g.getSnakeById(id)
		return snakeSnapshot, snakeSnapshot.Alive()
	})
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
	return g.UpdateGameSnapshotBoardState(nextBoardState), nil
}

func NewGameSnapshot(request *client.SnakeRequest) GameSnapshot {
	if request == nil {
		log.Println("Error: Request is nil")
		return nil
	}
	boardState := ConvertToBoardState(*request)

	rulesetName := request.Game.Ruleset.Name
	// log.Println("Creating game snapshot for ruleset:", rulesetName)

	ruleset := rules.NewRulesetBuilder().
		WithParams(ConvertRulesetSettingsToMap(request.Game.Ruleset.Settings)).
		WithSolo(len(request.Board.Snakes) < 2).
		NamedRuleset(rulesetName)

	if ruleset == nil {
		panic("Failed to create ruleset for request: " + rulesetName)
	}

	snakeStats := make(map[string]*snakeStatsImpl)
	for _, snake := range request.Board.Snakes {
		turnLastShouted := mo.TupleToOption(request.Turn, snake.Shout != "").OrElse(0)

		snakeStats[snake.ID] = &snakeStatsImpl{
			name:            snake.Name,
			lastShout:       snake.Shout,
			turnLastShouted: turnLastShouted,
		}
	}
	color := request.You.Customizations.Color

	allyIDs := lo.FilterMap(request.Board.Snakes, func(snake client.Snake, _ int) (string, bool) {
		return snake.ID, snake.Customizations.Color == color
	})

	opponentIDs := lo.FilterMap(request.Board.Snakes, func(snake client.Snake, _ int) (string, bool) {
		return snake.ID, snake.Customizations.Color != color
	})

	return &gameSnapshotImpl{
		gameID:      request.Game.ID,
		ruleset:     ruleset,
		boardState:  boardState,
		snakeStats:  snakeStats,
		yourID:      request.You.ID,
		allyIDs:     allyIDs,
		opponentIDs: opponentIDs,
	}
}

func (g *gameSnapshotImpl) UpdateGameSnapshotBoardState(newBoardState *rules.BoardState) GameSnapshot {
	if newBoardState == nil {
		panic("UpdateGameSnapshotBoardState: newBoardState is nil")
	}
	return &gameSnapshotImpl{
		gameID:      g.gameID,
		boardState:  newBoardState,
		ruleset:     g.ruleset,
		snakeStats:  g.snakeStats,
		yourID:      g.yourID,
		allyIDs:     g.allyIDs,
		opponentIDs: g.opponentIDs,
	}
}
