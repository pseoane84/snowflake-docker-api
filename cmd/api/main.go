package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"snowflake-docker-api/internal/db"
)

type App struct {
	DB *sql.DB
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Endpoint 1: proves live Snowflake connectivity
func (a *App) handleSFTime(w http.ResponseWriter, r *http.Request) {
	row := a.DB.QueryRow(`SELECT CURRENT_TIMESTAMP()`)
	var now time.Time
	if err := row.Scan(&now); err != nil {
		writeJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "snowflake_now": now.UTC().Format(time.RFC3339)})
}

// Endpoint 2: shows current session context (great for debugging roles/db/schema)
func (a *App) handleSFContext(w http.ResponseWriter, r *http.Request) {
	row := a.DB.QueryRow(`SELECT CURRENT_ACCOUNT(), CURRENT_USER(), CURRENT_ROLE(), CURRENT_WAREHOUSE(), CURRENT_DATABASE(), CURRENT_SCHEMA()`)
	var account, user, role, wh, dbName, schema string
	if err := row.Scan(&account, &user, &role, &wh, &dbName, &schema); err != nil {
		writeJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":        true,
		"account":   account,
		"user":      user,
		"role":      role,
		"warehouse": wh,
		"database":  dbName,
		"schema":    schema,
	})
}

// Endpoint 3: lists tables from INFORMATION_SCHEMA (live query)
func (a *App) handleSFMenuItems(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 25, 200)

	rows, err := a.DB.Query(`
		SELECT TABLE_SCHEMA, TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		ORDER BY TABLE_SCHEMA, TABLE_NAME
		LIMIT ?`, limit)
	if err != nil {
		writeJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	type rowOut struct {
		TableSchema string `json:"table_schema"`
		TableName   string `json:"table_name"`
	}

	out := []rowOut{}
	for rows.Next() {
		var r rowOut
		if err := rows.Scan(&r.TableSchema, &r.TableName); err != nil {
			writeJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		out = append(out, r)
	}

	writeJSON(w, 200, map[string]any{"ok": true, "count": len(out), "rows": out})
}

func parseLimit(raw string, def int, max int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

// HTMX: Home page
func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	// Serve the static HTML file
	http.ServeFile(w, r, "web/index.html")
}

// HTMX: Returns an HTML fragment (not JSON) for swapping into the page
func (a *App) handleHTMXTime(w http.ResponseWriter, r *http.Request) {
	var now string
	err := a.DB.QueryRow("SELECT TO_VARCHAR(CURRENT_TIMESTAMP())").Scan(&now)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(500)
		fmt.Fprintf(w, "<span style='color:red;'>Error: %s</span>", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<strong>Snowflake time:</strong> %s", now)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sfDB, err := db.Connect()
	if err != nil {
		log.Fatal("Snowflake connect failed: ", err)
	}
	defer sfDB.Close()

	app := &App{DB: sfDB}
	mux := http.NewServeMux()

	// Deliverable: 3 endpoints that connect to Snowflake
	mux.HandleFunc("/api/sf/time", app.handleSFTime)
	mux.HandleFunc("/api/sf/context", app.handleSFContext)
	mux.HandleFunc("/api/sf/menu-items", app.handleSFMenuItems)

	// HTMX demo endpoints
	mux.HandleFunc("/htmx/time", app.handleHTMXTime)
	mux.HandleFunc("/", app.handleHome)

	// Optional: JSON index of endpoints (useful for curl)
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"ok": true,
			"endpoints": []string{
				"/",
				"/htmx/time",
				"/api/sf/time",
				"/api/sf/context",
				"/api/sf/menu-items?limit=10",
			},
		})
	})

	log.Println("API listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
