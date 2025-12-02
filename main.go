package main

import (
	"log/slog"

	"github.com/webitel/meetings/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		slog.Error(err.Error())
		return
	}
}
