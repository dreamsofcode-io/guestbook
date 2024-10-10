package app

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/dreamsofcode-io/guestbook/internal/database"
	"github.com/dreamsofcode-io/guestbook/internal/middleware"
)

type App struct {
	logger *slog.Logger
	router *http.ServeMux
	db     *pgxpool.Pool
	rdb    *redis.Client
}

func New(logger *slog.Logger) *App {
	router := http.NewServeMux()

	redisAddr, exists := os.LookupEnv("REDIS_ADDR")
	if !exists {
		redisAddr = "localhost:6379"
	}

	app := &App{
		logger: logger,
		router: router,
		rdb: redis.NewClient(&redis.Options{
			Addr: redisAddr,
		}),
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	db, err := database.Connect(ctx, a.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	a.db = db

	tmpl := template.Must(template.New("").ParseGlob("./templates/*"))

	a.loadRoutes(tmpl)

	server := http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: middleware.Logging(a.logger, middleware.HandleBadCode(tmpl, a.router)),
	}

	done := make(chan struct{})
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("failed to listen and serve", slog.Any("error", err))
		}
		close(done)
	}()

	a.logger.Info("Server listening", slog.String("addr", ":8080"))
	select {
	case <-done:
		break
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		server.Shutdown(ctx)
		cancel()
	}

	return nil
}
