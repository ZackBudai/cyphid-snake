package main

import (
  "github.com/BattlesnakeOfficial/rules"
  "github.com/BattlesnakeOfficial/rules/client"
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
}

type gameSnapshotImpl struct {
  request    *client.SnakeRequest
  boardState *rules.BoardState
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

// SnakeSnapshot interface implementation

func (s *snakeSnapshotImpl) ID() string {
  return s.originalSnake.ID
}

func (s *snakeSnapshotImpl) Name() string {
  return s.originalSnake.Name
}

func (s *snakeSnapshotImpl) Health() int {
  if s.newSnake != nil {
    return s.newSnake.Health
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