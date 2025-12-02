package main

import (
	"log/slog"

	"github.com/webitel/web-meeting-backend/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		slog.Error(err.Error())
		return
	}
}
