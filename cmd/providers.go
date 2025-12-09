package cmd

import (
	"context"
	"fmt"
	"github.com/webitel/web-meeting-backend/infra/chat"
	"github.com/webitel/web-meeting-backend/infra/engine"
	"github.com/webitel/web-meeting-backend/infra/pubsub"
	sqlStore "github.com/webitel/web-meeting-backend/internal/store/sql"

	"github.com/webitel/web-meeting-backend/config"
	"github.com/webitel/web-meeting-backend/infra/consul"
	"github.com/webitel/web-meeting-backend/infra/encrypter"
	"github.com/webitel/web-meeting-backend/infra/grpc_srv"
	"github.com/webitel/web-meeting-backend/infra/sql/pgsql"
	"github.com/webitel/web-meeting-backend/internal/handler"
	"github.com/webitel/web-meeting-backend/internal/model"
	"github.com/webitel/web-meeting-backend/internal/service"
	"github.com/webitel/wlog"
	"go.uber.org/fx"
)

func ProvideLogger(cfg *config.Config, lc fx.Lifecycle) (*wlog.Logger, error) {
	logSettings := cfg.Log

	if !logSettings.Console && !logSettings.Otel && len(logSettings.File) == 0 {
		logSettings.Console = true
	}

	logConfig := &wlog.LoggerConfiguration{
		EnableConsole: logSettings.Console,
		ConsoleJson:   false,
		ConsoleLevel:  logSettings.Lvl,
	}

	if logSettings.File != "" {
		logConfig.FileLocation = logSettings.File
		logConfig.EnableFile = true
		logConfig.FileJson = true
		logConfig.FileLevel = logSettings.Lvl
	}

	l := wlog.NewLogger(logConfig)
	wlog.RedirectStdLog(l)
	wlog.InitGlobalLogger(l)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// Logger cleanup if needed
			return nil
		},
	})

	return l, nil
}

func ProvideGrpcServer(cfg *config.Config, l *wlog.Logger, lc fx.Lifecycle) (*grpc_srv.Server, error) {
	s, err := grpc_srv.New(cfg.Service.Address, l)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := s.Shutdown(); err != nil {
				l.Error(err.Error(), wlog.Err(err))
				return err
			}
			return nil
		},
	})

	return s, nil
}

func ProvideCluster(cfg *config.Config, srv *grpc_srv.Server, l *wlog.Logger, lc fx.Lifecycle) (*consul.Cluster, error) {
	c := consul.NewCluster(model.ServiceName, cfg.Service.Consul, l)
	host := srv.Host()

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return c.Start(cfg.Service.Id, host, srv.Port())
		},
		OnStop: func(ctx context.Context) error {
			c.Stop()
			return nil
		},
	})

	return c, nil
}

func ProvidePubSub(cfg *config.Config, l *wlog.Logger, lc fx.Lifecycle) (*pubsub.Manager, error) {

	ps, err := pubsub.New(l, cfg.Pubsub.Address)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return ps.Start()
		},
		OnStop: func(ctx context.Context) error {
			return ps.Shutdown()
		},
	})

	return ps, nil
}

func ProvideEncrypter(cfg *config.Config) (*encrypter.DataEncrypter, error) {
	if cfg.Service.SecretKey == "" {
		// Fallback або помилка. Для dev можна дефолтний, але краще помилку.
		// Для спрощення поки що повернемо помилку якщо пусто
		return nil, fmt.Errorf("service.secret_key is required")
	}
	return encrypter.New([]byte(cfg.Service.SecretKey))
}

func ProvideChat(cfg *config.Config, l *wlog.Logger, lc fx.Lifecycle) (*chat.Client, error) {
	cli, err := chat.NewClient(cfg.Service.Consul, l)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return cli.Close()
		},
	})

	return cli, nil
}

func ProvideEngine(cfg *config.Config, l *wlog.Logger, lc fx.Lifecycle) (*engine.Client, error) {
	cli, err := engine.NewClient(cfg.Service.Consul, l)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return cli.Close()
		},
	})

	return cli, nil
}

func ProvideContext() context.Context {
	return context.Background()
}

// ProvideMeetingStore створює сховище (SQL або Memory) в залежності від конфігурації
func ProvideMeetingStore(
	ctx context.Context,
	cfg *config.Config,
	log *wlog.Logger,
	lc fx.Lifecycle,
) (service.MeetingStore, error) {

	log.Info("Using SQL Meeting Store (PostgreSQL)")

	db, err := pgsql.New(ctx, cfg.SqlSettings.DSN, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cancel()
			log.Debug("Closing database connection")
			return db.Close()
		},
	})

	return sqlStore.NewMeetingStore(cancelCtx, db, log), nil
}

// ProvideMeetingService - адаптер для прив'язки service.MeetingService → handler.MeetingService
func ProvideMeetingService(impl *service.MeetingService) handler.MeetingService {
	return impl
}
