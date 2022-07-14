package modules

import (
	"context"
	"database/sql"
	"encoding/json"
	"matchmod/types"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

const newMatchOpCode = 999

type MatchHandler struct {
	Marshaler   *protojson.MarshalOptions
	Unmarshaler *protojson.UnmarshalOptions
}

func (m *MatchHandler) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &types.MatchState{
		EmptyTicks: 0,
		Presences:  map[string]runtime.Presence{},
		Status:     types.MatchStatusNotStarted,
	} // Define custom MatchState in the code as per your game's requirements
	tickRate := 20 // Call MatchLoop() every 20 times per second.
	label := ""    // Custom label that will be used to filter match listings.

	return state, tickRate, label
}

func (m *MatchHandler) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	MatchState, ok := state.(*types.MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
	}

	for i := 0; i < len(presences); i++ {
		MatchState.Presences[presences[i].GetSessionId()] = presences[i]
	}

	return MatchState
}

func (m *MatchHandler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	result := true

	return state, result, ""
}

func (m *MatchHandler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	matchState, ok := state.(*types.MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
	}

	for i := 0; i < len(presences); i++ {
		delete(matchState.Presences, presences[i].GetSessionId())
	}

	return matchState
}

func (m *MatchHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {

	matchState, ok := state.(*types.MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
	}

	if len(matchState.Presences) == 0 {
		matchState.EmptyTicks++
	}

	if matchState.EmptyTicks > 100 {
		return nil
	}

	return matchState
}

func (m *MatchHandler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}

func (m *MatchHandler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	logger.Debug("match will terminate in %d seconds", graceSeconds)

	var matchId string

	// Find an existing match for the remaining connected presences to join
	limit := 1
	authoritative := true
	label := ""
	minSize := 2
	maxSize := 4
	query := "*"
	availableMatches, err := nk.MatchList(ctx, limit, authoritative, label, &minSize, &maxSize, query)
	if err != nil {
		logger.Error("error listing matches", err)
		return nil
	}

	if len(availableMatches) > 0 {
		matchId = availableMatches[0].MatchId
	} else {
		// No available matches, create a new match instead
		matchId, err = nk.MatchCreate(ctx, "match", nil)
		if err != nil {
			logger.Error("error creating match", err)
			return nil
		}
	}

	// Broadcast the new match id to all remaining connected presences
	data := map[string]string{
		matchId: matchId,
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		logger.Error("error marshaling new match message")
		return nil
	}

	dispatcher.BroadcastMessage(newMatchOpCode, dataJson, nil, nil, true)

	return state
}
