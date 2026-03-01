package web

import (
	"context"
	"crabspy"
	"crabspy/sql/sqlcgen"
	"encoding/json"
	"time"
)

type GameState struct {
	RoomID          int64
	SpyID           int64
	Location        crabspy.Location
	StartedAt       int64
	TimerDuration   int64
	Paused          bool
	PausedID        int64 // 0 = nobody
	AccusedID       int64 // 0 = nobody
	Events          []sqlcgen.GameEvent
	HasPaused       map[int64]bool
	HasAccused      map[int64]bool
	GuessedLocation string
}

func BuildGameState(game sqlcgen.Game, events []sqlcgen.GameEvent) GameState {
	var location crabspy.Location
	for _, l := range crabspy.Locations {
		if l.Title == game.Location {
			location = l
		}
	}
	state := GameState{
		RoomID:        game.RoomID,
		SpyID:         game.SpyID,
		Location:      location,
		StartedAt:     game.StartedAt,
		TimerDuration: game.TimerDuration,
		Events:        events,
	}

	state.HasPaused = make(map[int64]bool)
	state.HasAccused = make(map[int64]bool)

	for _, e := range events {
		switch e.EventType {
		case "paused":
			state.Paused = true
			state.PausedID = e.UserID
			state.HasPaused[e.UserID] = true
		case "unpaused":
			state.Paused = false
			state.PausedID = 0
			state.AccusedID = 0
		case "accused":
			if e.TargetID.Valid && e.TargetID.Int64 != e.UserID {
				state.AccusedID = e.TargetID.Int64
				state.HasAccused[e.UserID] = true
			}
		case "location_guessed":
			if e.Metadata.Valid {
				var m map[string]string
				json.Unmarshal([]byte(e.Metadata.String), &m)
				state.GuessedLocation = m["location"]
			}
		}
	}
	return state
}

func getGameState(ctx context.Context, q *sqlcgen.Queries, roomID int64) (sqlcgen.Game, GameState, error) {
	game, err := q.GetCurrentGame(ctx, roomID)
	if err != nil {
		return game, GameState{}, err
	}
	events, _ := q.GetGameEvents(ctx, game.ID)
	return game, BuildGameState(game, events), nil
}

func (s GameState) TimerRemaining() int64 {
	var elapsed int64
	unpauseStart := s.StartedAt

	for _, e := range s.Events {
		switch e.EventType {
		case "paused":
			elapsed += e.CreatedAt - unpauseStart
		case "unpaused":
			unpauseStart = e.CreatedAt
		}
	}

	if !s.Paused {
		elapsed += time.Now().Unix() - unpauseStart
	}

	remaining := s.TimerDuration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}
