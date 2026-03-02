package web

import (
	"context"
	"crabspy"
	"crabspy/sql/sqlcgen"
	"encoding/json"
	"time"
)

type VoteOutcome int

const (
	VoteIncomplete    VoteOutcome = iota // not everyone voted yet
	VoteSpyCaught                        // unanimous + correct
	VoteWrongPlayer                      // unanimous + wrong person
	VoteNoConsensus                      // everyone voted, not unanimous mid-game
	VoteTimeupSpyWins                    // everyone voted, not unanimous timer expired
)

type GameState struct {
	RoomID          int64
	SpyID           int64
	Location        crabspy.Location
	LocationPool    []string
	StartedAt       int64
	TimerDuration   int64
	Paused          bool
	PausedID        int64 // 0 = nobody
	AccusedID       int64 // 0 = nobody
	Events          []sqlcgen.GameEvent
	HasPaused       map[int64]bool
	HasAccused      map[int64]bool
	GuessedLocation string
	Votes           map[int64]string
}

func BuildGameState(game sqlcgen.Game, events []sqlcgen.GameEvent) GameState {
	var location crabspy.Location
	for _, l := range crabspy.Locations {
		if l.Title == game.Location {
			location = l
		}
	}
	var lp []string
	json.Unmarshal([]byte(game.LocationPool), &lp)
	state := GameState{
		RoomID:        game.RoomID,
		SpyID:         game.SpyID,
		Location:      location,
		StartedAt:     game.StartedAt,
		TimerDuration: game.TimerDuration,
		LocationPool:  lp,
		Events:        events,
	}

	state.HasPaused = make(map[int64]bool)
	state.HasAccused = make(map[int64]bool)
	state.Votes = make(map[int64]string)

	for _, e := range events {
		switch e.EventType {
		case "paused":
			state.Paused = true
			state.PausedID = e.UserID
			state.HasPaused[e.UserID] = true
			//Clear votes on pause
			state.Votes = make(map[int64]string)
		case "unpaused":
			state.Paused = false
			state.PausedID = 0
			state.AccusedID = 0
			//Clear votes on unpause
			state.Votes = make(map[int64]string)
		case "accused":
			if e.TargetID.Valid && e.TargetID.Int64 != e.UserID {
				state.AccusedID = e.TargetID.Int64
				state.HasAccused[e.UserID] = true
				if e.Metadata.Valid {
					var m map[string]string
					json.Unmarshal([]byte(e.Metadata.String), &m)
					state.Votes[e.UserID] = m["vote"]
				}
			}
		case "voted":
			if e.Metadata.Valid {
				var m map[string]string
				json.Unmarshal([]byte(e.Metadata.String), &m)
				state.Votes[e.UserID] = m["vote"]
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

func (g *GameState) IsUnanimousSpy(memberCount int) bool {
	eligible := memberCount - 1
	spyVotes := 0
	for _, v := range g.Votes {
		if v == "spy" {
			spyVotes++
		}
	}
	// If spy votes for himself, still call it unanimous
	return spyVotes >= eligible
}

func (g *GameState) ResolveVote(memberCount int) VoteOutcome {
	if len(g.Votes) < memberCount-1 {
		return VoteIncomplete
	}
	if g.IsUnanimousSpy(memberCount) && g.AccusedID != 0 {
		if g.AccusedID == g.SpyID {
			return VoteSpyCaught
		}
		return VoteWrongPlayer
	}
	if g.TimerRemaining() == 0 {
		return VoteTimeupSpyWins
	}
	return VoteNoConsensus
}
