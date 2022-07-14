package types

import "github.com/heroiclabs/nakama-common/runtime"

const (
	MatchStatusNotStarted = 0
	MatchStatusRunning    = 1
	MatchStatusFinished   = 3
)

type MatchState struct {
	Presences  map[string]runtime.Presence
	EmptyTicks int
	Status     int
}
