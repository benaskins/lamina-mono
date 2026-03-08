package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/benaskins/axon"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var requestCount atomic.Int64

func main() {
	port := "9900"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://aurelia:aurelia@localhost:5432/aurelia?sslmode=disable"
	}

	schema := fmt.Sprintf("loadtest_%s", port)
	db, err := axon.OpenDB(dsn, schema)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	maxConns := 25
	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if n, err := fmt.Sscanf(v, "%d", &maxConns); n == 1 && err == nil {
			// use parsed value
		}
	}
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns / 2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := setupTable(db); err != nil {
		slog.Error("failed to create table", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("GET /api/echo", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		axon.WriteJSON(w, http.StatusOK, map[string]any{
			"message": r.URL.Query().Get("msg"),
			"time":    time.Now().UnixMilli(),
		})
	})

	// Write path: insert a row, return the ID
	mux.HandleFunc("POST /api/items", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			axon.WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if body.Name == "" {
			body.Name = "item"
		}

		var id int64
		err := db.QueryRowContext(r.Context(),
			"INSERT INTO items (name) VALUES ($1) RETURNING id", body.Name,
		).Scan(&id)
		if err != nil {
			axon.WriteError(w, http.StatusInternalServerError, "insert failed")
			return
		}

		axon.WriteJSON(w, http.StatusCreated, map[string]any{
			"id":   id,
			"name": body.Name,
			"port": port,
		})
	})

	// Read path: fetch recent items
	mux.HandleFunc("GET /api/items", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		rows, err := db.QueryContext(r.Context(),
			"SELECT id, name, created_at FROM items ORDER BY id DESC LIMIT 20",
		)
		if err != nil {
			axon.WriteError(w, http.StatusInternalServerError, "query failed")
			return
		}
		defer rows.Close()

		type item struct {
			ID        int64     `json:"id"`
			Name      string    `json:"name"`
			CreatedAt time.Time `json:"created_at"`
		}
		var items []item
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.ID, &it.Name, &it.CreatedAt); err != nil {
				continue
			}
			items = append(items, it)
		}

		axon.WriteJSON(w, http.StatusOK, map[string]any{
			"items": items,
			"count": len(items),
			"port":  port,
		})
	})

	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		var itemCount int64
		db.QueryRow("SELECT count(*) FROM items").Scan(&itemCount)

		var poolStats struct {
			open, idle, inUse int
		}
		s := db.Stats()
		poolStats.open = s.OpenConnections
		poolStats.idle = s.Idle
		poolStats.inUse = s.InUse

		axon.WriteJSON(w, http.StatusOK, map[string]any{
			"requests":   requestCount.Load(),
			"items":      itemCount,
			"port":       port,
			"pool_open":  poolStats.open,
			"pool_idle":  poolStats.idle,
			"pool_inuse": poolStats.inUse,
		})
	})

	slog.Info("starting loadtest service", "port", port, "schema", schema, "max_conns", maxConns)

	axon.ListenAndServe(port, mux,
		axon.WithDrainTimeout(5*time.Second),
	)
}

func setupTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}
