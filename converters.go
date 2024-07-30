package main

import (
	"github.com/BattlesnakeOfficial/rules"
)

func (request SnakeRequest) ConvertToBoardState() *rules.BoardState {
	// Create a new BoardState with the same dimensions as the GameState's Board
	bs := rules.NewBoardState(request.Board.Width, request.Board.Height)

	// Set the turn
	bs.Turn = request.Turn

	// Convert food
	for _, food := range request.Board.Food {
		bs.Food = append(bs.Food, rules.Point{
			X: food.X,
			Y: food.Y,
		})
	}

	// Convert hazards
	for _, hazard := range request.Board.Hazards {
		bs.Hazards = append(bs.Hazards, rules.Point{
			X: hazard.X,
			Y: hazard.Y,
		})
	}

	// Convert snakes
	for _, snake := range request.Board.Snakes {
		newSnake := rules.Snake{
			ID:     snake.ID,
			Health: snake.Health,
		}

		// Convert snake body
		for _, bodyPart := range snake.Body {
			newSnake.Body = append(newSnake.Body, rules.Point{
				X: bodyPart.X,
				Y: bodyPart.Y,
			})
		}

		bs.Snakes = append(bs.Snakes, newSnake)
	}

	return bs
}
