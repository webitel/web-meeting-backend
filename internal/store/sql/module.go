package sql

import "go.uber.org/fx"

var Module = fx.Module("store",
	fx.Provide(NewMeetingStore),
)
