package web

import "crabspy/sql/sqlcgen"

type GameState struct {
	RoomID         int64
	SpyID          int64
	Location       string
	StartedAt      int64
	TimerDuration  int64
	Paused         bool
	PausedID       *int64
	AccusedID      *int64
	TimerRemaining int64
	Events         []GameEvent
}

type GameEvent struct {
	RoomID    int64
	SpyID     int64
	Location  string
	EventType string
	CreatedAt string
}

func buildGameState(game sqlcgen.Game, events []GameEvent) GameState {
	//get base game state and add event info
	state := GameState{
		RoomID:         game.RoomID,
		SpyID:          game.SpyID,
		Location:       game.Location,
		StartedAt:      game.StartedAt,
		TimerDuration:  game.TimerDuration,
		TimerRemaining: game.TimerRemaining,
		Events:         events,
	}

}
