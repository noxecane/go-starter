package config

import (
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/go-redis/redis/v8"
	"github.com/tsaron/anansi"
	"github.com/tsaron/anansi/tokens"
)

type App struct {
	Env    *Env
	DB     *pg.DB
	Redis  *redis.Client
	Tokens *tokens.Store
	Auth   *anansi.SessionStore
}

func HealthChecker(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		if err := app.DB.Ping(r.Context()); err != nil {
			http.Error(w, "Could not reach postgres", http.StatusInternalServerError)
			return
		}

		if _, err := app.Redis.Ping(r.Context()).Result(); err != nil {
			http.Error(w, "Could not reach redis", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		// we don't have a plan for when writes fail
		_, _ = w.Write([]byte("Up and Running!"))
	}
}
