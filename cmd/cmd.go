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
		Name:     "web-meeting",
		Usage:    "web-meeting in the Webitel",
		Compiled: time.Now(),
		Action: func(c *cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			apiCmd(cfg),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Category:    "observability/logging",
				Usage:       "application log level",
				EnvVars:     []string{"LOG_LVL"},
				Value:       "debug",
				Destination: &cfg.Log.Lvl,
				Aliases:     []string{"l"},
			},
			&cli.BoolFlag{
				Name:        "log-json",
				Category:    "observability/logging",
				Usage:       "application log json",
				Value:       false,
				EnvVars:     []string{"LOG_JSON"},
				Destination: &cfg.Log.Json,
			},
			&cli.BoolFlag{
				Name:        "log-otel",
				Category:    "observability/logging",
				Usage:       "application log OTEL",
				Value:       false,
				EnvVars:     []string{"LOG_OTEL"},
				Destination: &cfg.Log.Otel,
			},
			&cli.BoolFlag{
				Name:        "log-console",
				Category:    "observability/logging",
				Usage:       "application log stdout",
				Value:       true,
				EnvVars:     []string{"LOG_CONSOLE"},
				Destination: &cfg.Log.Console,
			},
			&cli.StringFlag{
				Name:        "log-file",
				Category:    "observability/logging",
				Usage:       "application log file",
				Value:       "",
				EnvVars:     []string{"LOG_FILE"},
				Destination: &cfg.Log.File,
			},
			&cli.StringFlag{
				Name:        "data-encrypter",
				Category:    "crypto",
				Usage:       "Secret key",
				EnvVars:     []string{"SECRET_KEY"},
				Value:       "MY_SECRET_KEY", //
				Destination: &cfg.Service.SecretKey,
			},
			&cli.StringFlag{
				Name:        "pubsub",
				Category:    "service/pubsub",
				Usage:       "publish/subscribe rabbitmq broker connection string",
				Value:       "amqp://admin:admin@127.0.0.1:5672/",
				Destination: &cfg.Pubsub.Address,
				Aliases:     []string{"p"},
				EnvVars:     []string{"PUBSUB"},
			},
		},
	}

	if err := def.Run(os.Args); err != nil {
		return err
	}

	return nil
}
