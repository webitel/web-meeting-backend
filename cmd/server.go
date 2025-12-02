package cmd

import (
	"context"
	"log/slog"

	"github.com/urfave/cli/v2"
	"github.com/webitel/meetings/config"
	"github.com/webitel/meetings/infra/auth"
	wauth "github.com/webitel/meetings/infra/auth/manager/webitel"
	"github.com/webitel/meetings/infra/consul"
	server "github.com/webitel/meetings/infra/server/grpc"
	"github.com/webitel/webitel-go-kit/infra/pubsub/rabbitmq"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	// Import OTEL SDK components
	slogutil "github.com/webitel/webitel-go-kit/infra/otel/log/bridge/slog"
	otelsdk "github.com/webitel/webitel-go-kit/infra/otel/sdk"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/log/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/log/stdout"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/metric/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/metric/stdout"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/trace/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/trace/stdout"
)

const (
	serviceName = "meetings.service"
)

var (
	service = resource.NewSchemaless(
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion("0.1.0"),
		semconv.ServiceInstanceID("example"),
		semconv.ServiceNamespace("webitel"),
	)

	verbose slog.LevelVar
)

func serverCmd(conf *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"a"},
		Usage:   "",
		Flags:   serverFlags(conf),
		Action: func(c *cli.Context) error {
			// Start the meetings service

			return nil
		},
	}
}

func serverFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "bind-address",
			Category:    "server",
			Usage:       "address that should be bound to for internal cluster communications",
			Value:       "127.0.0.1:10041",
			Destination: &cfg.ServerAddress,
			Aliases:     []string{"b"},
			EnvVars:     []string{"BIND_ADDRESS"},
		},
		&cli.StringFlag{
			Name:     "consul-address",
			Category: "consul",
			Usage:    "consul address (host:port)",
			Value:    "127.0.0.1:8500",
			EnvVars:  []string{"CONSUL_ADDRESS"},
		},
		&cli.StringFlag{
			Name:     "consul-token",
			Category: "consul",
			Usage:    "consul ACL token",
			EnvVars:  []string{"CONSUL_TOKEN"},
		},
	}
}

type application struct {
	config *config.Config

	authManager auth.Manager
	server      *server.Server
	broker      *rabbitmq.Broker
	registry    *consul.Consul
}

func configureOTEL() (otelsdk.ShutdownFunc, error) {
	ctx := context.Background()
	shutdown, err := otelsdk.Configure(ctx,
		otelsdk.WithResource(service),
		otelsdk.WithLogBridge(func() {
			// Just for example ...
			// Redirect slog.Default().Handler() to
			// otel/log/global.LoggerProvider with slog.Level filter
			stdlog := slog.New(
				slogutil.WithLevel(
					// front: otelslog.Handler level filter
					&verbose,
					// back: otel/log/global.Logger("slog")
					otelslog.NewHandler("slog"),
				),
			)

			slog.SetDefault(stdlog)

		}),
	)
	if err != nil {
		return nil, err
	}
	return shutdown, nil
}

func NewApplication(cfg *config.Config) (*application, error) {
	otelShutdown, err := configureOTEL()
	if err != nil {
		return nil, err
	}
	defer otelShutdown(context.Background())

	return &application{config: cfg}, nil
}
