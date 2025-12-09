package cmd

import (
	"github.com/webitel/web-meeting-backend/config"
	"github.com/webitel/web-meeting-backend/internal/handler"
	"github.com/webitel/web-meeting-backend/internal/service"
	"go.uber.org/fx"
)

func NewApp(cfg *config.Config) *fx.App {
	return fx.New(
		// Конфігурація та контекст
		fx.Supply(cfg),
		fx.Provide(ProvideContext),
		fx.Provide(ProvidePubSub),
		fx.Provide(ProvideEncrypter),

		// Infrastructure providers
		fx.Provide(ProvideLogger),
		fx.Provide(ProvideGrpcServer),
		fx.Provide(ProvideCluster),
		fx.Provide(ProvideChat),
		fx.Provide(ProvideEngine),

		// Адаптери для прив'язки інтерфейсів
		fx.Provide(ProvideMeetingStore),   // store.MeetingStoreImpl → service.MeetingStore
		fx.Provide(ProvideMeetingService), // service.MeetingService → handler.MeetingService

		// Business logic modules
		service.Module,
		handler.Module,

		// Invoke startup functions
		fx.Invoke(StartGrpcServer),
		fx.Invoke(RegisterHandlers),
		fx.Invoke(EnsureCluster),

		// fx налаштування
		fx.NopLogger, // Вимикаємо fx логи, використовуємо наш logger
	)
}
