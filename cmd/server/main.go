package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"summer-camp-scheduler/internal/database"
	"summer-camp-scheduler/internal/handlers"
)

func main() {
	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "3306")
	user := envOrDefault("DB_USER", "root")
	pass := envOrDefault("DB_PASSWORD", "campscheduler")
	dbName := envOrDefault("DB_NAME", "camp")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbName)

	var db *sql.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			break
		}
		log.Printf("Waiting for database... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	store := database.NewStore(db)
	h := handlers.New(store)

	mux := http.NewServeMux()
	h.Register(mux)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("Summer Camp Scheduler running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
