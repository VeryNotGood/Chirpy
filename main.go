package main

import (
	"VeryNotGood/Chirpy/internal/database"
	"VeryNotGood/Chirpy/middleware"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	const port = "8080"

	dbURL := os.Getenv("DB_URL")
	// dbURL := "postgres://shawnbelmore:@localhost:5432/chirpy"

	fmt.Println(dbURL)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database... %v", err)
	}

	dbErr := db.Ping()
	if dbErr != nil {
		log.Fatalf("Database Error: %v", dbErr)
	}

	apiCfg := new(middleware.ApiConfig)

	apiCfg.DBQuery = database.New(db)

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetHandler)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.ValidateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.AddUser)

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("Server running on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
