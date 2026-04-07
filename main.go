package main

import _ "github.com/lib/pq"

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/SaschaRunge/chirpy/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("unable to open database: %s", err)
		return
	}

	cfg := apiConfig{}
	cfg.dbQueries = database.New(db)
	cfg.platform = os.Getenv("PLATFORM")

	mux := http.NewServeMux()
	mux.Handle(
		"/app/",
		http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))),
	)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerJsonResponse)

	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)

	mux.HandleFunc("GET /admin/metrics", cfg.handlerReturnFileServerHits)
	mux.HandleFunc("POST /admin/reset", cfg.handlerResetUsers)

	server := http.Server{
		Addr:                         ":8080",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
	}

	err = server.ListenAndServe()

	fmt.Println(err)
}
