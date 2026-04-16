package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DBPing возвращает handler для проверки доступности PostgreSQL.
func DBPing(db *sql.DB, log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Error("PostgreSQL ping failed", zap.Error(err))

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("pong")); err != nil {
			return
		}
	}
}
