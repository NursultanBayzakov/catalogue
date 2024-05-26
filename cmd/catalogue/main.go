package main

import (
	"catalogue-service/config"
	"catalogue-service/internal/app"
	"catalogue-service/internal/sl"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local" // локальный запуск. Используем удобный для консоли TextHandler и уровень логирования Debug (будем выводить все сообщения).
	envDev   = "dev"   // удаленный dev-сервер. Уровень логирования тот же, но формат вывода — JSON, удобный для систем сбора логов вроде Kibana или Grafana Loki.
	envProd  = "prod"  // продакшен. Повышаем уровень логирования до Info, чтобы не выводить дебаг-логи в проде. То есть мы будем получать сообщения только с уровнем Info или Error.
)

func main() {
	cfg := config.LoadConfig()
	log := setupLogger(cfg.Env)
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTtl)

	go func() {
		application.GRPCServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()
	log.Info("Catalogue service gracefully stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := sl.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
