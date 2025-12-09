package cmd

import (
	"context"
	"fmt"
	"github.com/webitel/web-meeting-backend/infra/grpc_srv"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/webitel/web-meeting-backend/config"
	"github.com/webitel/web-meeting-backend/infra/consul"
	"github.com/webitel/web-meeting-backend/internal/handler"
	"github.com/webitel/wlog"
	"go.uber.org/fx"
)

// StartGrpcServer запускає gRPC сервер
func StartGrpcServer(lc fx.Lifecycle, srv *grpc_srv.Server, log *wlog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info(fmt.Sprintf("listen grpc %s:%d", srv.Host(), srv.Port()))
				if err := srv.Listen(); err != nil {
					log.Error("grpc server error", wlog.Err(err))
				}
			}()
			return nil
		},
	})
}

func RegisterHandlers(_ *handler.MeetingHandler, _ *handler.CallsHandler) {
	// Handlers автоматично реєструються в своїх конструкторах
}

func EnsureCluster(_ *consul.Cluster) {
	// Cluster автоматично реєструється через lifecycle hooks
}

func apiCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"a"},
		Usage:   "Start web-meeting-backend server",
		Flags:   apiFlags(cfg),
		Action: func(c *cli.Context) error {
			// Створюємо fx.App з усіма залежностями
			app := NewApp(cfg)

			if err := app.Start(c.Context); err != nil {
				wlog.Error("failed to start application", wlog.Err(err))
				return err
			}

			interruptChan := make(chan os.Signal, 1)
			signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-interruptChan

			wlog.Info("shutting down gracefully...")
			if err := app.Stop(context.Background()); err != nil {
				wlog.Error("error during shutdown", wlog.Err(err))
				return err
			}

			wlog.Info("application stopped")
			return nil
		},
	}
}

func apiFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "service-id",
			Category:    "server",
			Usage:       "service id ",
			Value:       "1",
			Destination: &cfg.Service.Id,
			Aliases:     []string{"i"},
			EnvVars:     []string{"ID"},
		},
		&cli.StringFlag{
			Name:        "bind-address",
			Category:    "server",
			Usage:       "address that should be bound to for internal cluster communications",
			Value:       "localhost:50011",
			Destination: &cfg.Service.Address,
			Aliases:     []string{"b"},
			EnvVars:     []string{"BIND_ADDRESS"},
		},
		&cli.StringFlag{
			Name:        "consul-discovery",
			Category:    "server",
			Usage:       "service discovery address",
			Value:       "127.0.0.1:8500",
			Destination: &cfg.Service.Consul,
			Aliases:     []string{"c"},
			EnvVars:     []string{"CONSUL"},
		},
		&cli.StringFlag{
			Name:        "postgresql-dsn",
			Category:    "database",
			Usage:       "Postgres connection string",
			EnvVars:     []string{"DATA_SOURCE"},
			Value:       "", // postgres://postgres:postgres@localhost:5432/webitel?sslmode=disable
			Destination: &cfg.SqlSettings.DSN,
		},
	}
}
