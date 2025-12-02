package cmd

import (
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/webitel/web-meeting-backend/config"
)

func Run() error {
	cfg := &config.Config{}
	def := &cli.App{
		Name:     "meetings-service",
		Usage:    "Video meetings manager in the Webitel",
		Compiled: time.Now(),
		Commands: []*cli.Command{
			serverCmd(cfg),
		},
	}

	if err := def.Run(os.Args); err != nil {
		return err
	}

	return nil
}
