package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	idLen   = 22
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var db *sql.DB

func generateID() (string, error) {
	b := make([]byte, idLen)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

func parseTTL(ttl string) (time.Duration, bool) {
	if ttl == "once" {
		return 24 * time.Hour, true // one-time shares expire after 24h if not viewed
	}
	switch ttl {
	case "15m":
		return 15 * time.Minute, false
	case "1h":
		return time.Hour, false
	case "6h":
		return 6 * time.Hour, false
	case "1d":
		return 24 * time.Hour, false
	case "7d":
		return 7 * 24 * time.Hour, false
	default:
		return time.Hour, false
	}
}

type createRequest struct {
	Tool  string          `json:"tool"`
	State json.RawMessage `json:"state"`
	TTL   string          `json:"ttl"`
}

type createResponse struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

type shareResponse struct {
	Tool      string          `json:"tool"`
	State     json.RawMessage `json:"state"`
	CreatedAt string          `json:"created_at"`
	ExpiresAt string          `json:"expires_at"`
	OneTime   bool            `json:"one_time"`
}

func handleCreateShare(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Tool == "" || len(req.State) == 0 {
		http.Error(w, "tool and state are required", http.StatusBadRequest)
		return
	}

	// Limit state size to 1MB
	if len(req.State) > 1<<20 {
		http.Error(w, "state too large", http.StatusRequestEntityTooLarge)
		return
	}

	id, err := generateID()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	ttlDuration, oneTime := parseTTL(req.TTL)
	expiresAt := time.Now().Add(ttlDuration)

	_, err = db.Exec(
		`INSERT INTO shares (id, tool, state, one_time, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, req.Tool, string(req.State), oneTime, time.Now().UTC().Format(time.RFC3339), expiresAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		log.Printf("error creating share: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createResponse{
		ID:        id,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	})
}

func handleGetShare(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "invalid share id", http.StatusBadRequest)
		return
	}

	var tool, state, createdAt, expiresAt string
	var oneTime bool

	err := db.QueryRow(
		`SELECT tool, state, one_time, created_at, expires_at FROM shares WHERE id = ?`, id,
	).Scan(&tool, &state, &oneTime, &createdAt, &expiresAt)

	if err == sql.ErrNoRows {
		http.Error(w, "share not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error fetching share: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Check expiration
	exp, _ := time.Parse(time.RFC3339, expiresAt)
	if time.Now().UTC().After(exp) {
		db.Exec(`DELETE FROM shares WHERE id = ?`, id)
		http.Error(w, "share not found", http.StatusNotFound)
		return
	}

	// Delete one-time shares after retrieval
	if oneTime {
		db.Exec(`DELETE FROM shares WHERE id = ?`, id)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shareResponse{
		Tool:      tool,
		State:     json.RawMessage(state),
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		OneTime:   oneTime,
	})
}

func cleanupExpired() {
	for {
		time.Sleep(60 * time.Second)
		result, err := db.Exec(`DELETE FROM shares WHERE expires_at < ?`, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			log.Printf("cleanup error: %v", err)
			continue
		}
		if n, _ := result.RowsAffected(); n > 0 {
			log.Printf("cleaned up %d expired shares", n)
		}
	}
}

func main() {
	distDir := os.Getenv("DIST_DIR")
	if distDir == "" {
		distDir = "../dist"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Ensure data directory exists
	os.MkdirAll(dataDir, 0755)

	dbPath := filepath.Join(dataDir, "shares.db")
	var err error
	db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS shares (
		id TEXT PRIMARY KEY,
		tool TEXT NOT NULL,
		state TEXT NOT NULL,
		one_time BOOLEAN NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		expires_at TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// Index for cleanup
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expires_at ON shares(expires_at)`)

	// Start cleanup goroutine
	go cleanupExpired()

	// API routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/share", handleCreateShare)
	mux.HandleFunc("GET /api/share/{id}", handleGetShare)

	// Static file serving with SPA fallback
	distAbs, _ := filepath.Abs(distDir)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := filepath.Join(distAbs, filepath.Clean(r.URL.Path))

		// Prevent directory traversal
		if !strings.HasPrefix(path, distAbs) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Check if file exists
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}

		// Check if it's a directory with index.html
		if err == nil && info.IsDir() {
			indexPath := filepath.Join(path, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// SPA fallback: serve index.html for all unmatched routes
		indexPath := filepath.Join(distAbs, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, indexPath)
	})

	log.Printf("serving on %s (dist: %s, data: %s)", addr, distAbs, dataDir)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
