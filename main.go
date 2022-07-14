package main

import (
	"context"
	"database/sql"
	"matchmod/modules"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Hello World!")

	marshaler := &protojson.MarshalOptions{
		UseEnumNumbers: true,
	}
	unmarshaler := &protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	if err := initializer.RegisterMatch("", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &modules.MatchHandler{
			Marshaler:   marshaler,
			Unmarshaler: unmarshaler,
		}, nil
	}); err != nil {
		logger.Error("unable to register: %v", err)
		return err
	}

	return nil
}
