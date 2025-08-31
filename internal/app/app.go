package app

import (
	"context"
	"github.com/ShlykovPavel/http_service_on_chi_template/internal/config"
	"github.com/ShlykovPavel/http_service_on_chi_template/internal/lib/api/middlewares"
	validators "github.com/ShlykovPavel/http_service_on_chi_template/internal/lib/api/validator"
	"github.com/ShlykovPavel/http_service_on_chi_template/internal/storage/database"
	"github.com/ShlykovPavel/http_service_on_chi_template/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App Структура приложения. Включает в себя все необходимые элементы для запуска приложения. (в последствии сюда можно докинуть gRPC итп)
type App struct {
	HTTPServer *http.Server
	logger     *slog.Logger
	cfg        *config.Config
}

// NewApp создаёт экземпляр приложения, инициализируя все зависимости:
// - Подключение к БД (пул соединений).
// - Репозитории для работы с данными.
// - Настройку роутера и HTTP-сервера.
func NewApp(logger *slog.Logger, cfg *config.Config) *App {

	dbConfig := database.DbConfig{
		DbName:              cfg.DbName,
		DbUser:              cfg.DbUser,
		DbPassword:          cfg.DbPassword,
		DbHost:              cfg.DbHost,
		DbPort:              cfg.DbPort,
		DbMaxConnections:    cfg.DbMaxConnections,
		DbMinConnections:    cfg.DbMinConnections,
		DbMaxConnLifetime:   cfg.DbMaxConnLifetime,
		DbMaxConnIdleTime:   cfg.DbMaxConnIdleTime,
		DbHealthCheckPeriod: cfg.DbHealthCheckPeriod,
	}
	//Инициализация метрик Prometheus
	metricses := metrics.InitMetrics()
	// Инициализация пул соединений БД
	poll, err := database.CreatePool(context.Background(), &dbConfig, logger)
	if err != nil {
		logger.Error("Failed to create database pool", "error", err)
		os.Exit(1)
	}
	// Мониторинг пула для передачи в метрики
	database.MonitorPool(context.Background(), poll, metricses)
	metricsMiddleware := middlewares.PrometheusMiddleware(metricses)
	// Инициализация экземпляра валидатора
	if err = validators.InitValidator(); err != nil {
		logger.Error("Failed to initialize validator", "error", err)
	}

	// TODO Инициализация реализация репозитория БД

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Heartbeat("/health"))
	router.Use(metricsMiddleware)

	router.Route("/api/v1", func(apiRouter chi.Router) {
		apiRouter.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/swagger/doc.json"),
		))

		apiRouter.Handle("/metrics", promhttp.Handler())
		// TODO Зарегистрировать http поинты

	})

	// Run server
	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: cfg.ServerTimeout,
		WriteTimeout:      cfg.ServerTimeout,
	}
	return &App{cfg: cfg, logger: logger, HTTPServer: srv}
}

// Run запускает HTTP-сервер и ожидает сигналов для graceful shutdown.
// Это позволяет добавить в будущем другие подсистемы (например, gRPC), вызывая их Run в горутинах.
func (a *App) Run() {
	a.logger.Info("Starting HTTP server", slog.String("address", a.cfg.Address))

	// Запуск сервера в горутине для возможности graceful shutdown
	go func() {
		if err := a.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Failed to start server", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Ожидание сигналов для shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.logger.Info("Shutting down server...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Можно вынести в config
	defer cancel()
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown", "error", err.Error())
		os.Exit(1)
	}

	a.logger.Info("Server stopped")
}
