package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dreamsofcode-io/guestbook/internal/database"
	"github.com/dreamsofcode-io/guestbook/internal/middleware"
)

type App struct {
	logger *slog.Logger
	router *http.ServeMux
	db     *pgxpool.Pool
}

func New(logger *slog.Logger) *App {
	router := http.NewServeMux()

	app := &App{
		logger: logger,
		router: router,
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	db, err := database.Connect(ctx, a.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	a.db = db

	a.loadRoutes()

	server := http.Server{
		Addr:    ":8080",
		Handler: middleware.Logging(a.logger, a.router),
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
