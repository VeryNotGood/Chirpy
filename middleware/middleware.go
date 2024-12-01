package middleware

import (
	"VeryNotGood/Chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type ApiConfig struct {
	FileServerHits atomic.Int32
	DBQuery        *database.Queries
}

type parameters struct {
	Body string `json:"body"`
}

type userReq struct {
	Email string `json:"email"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type validResponse struct {
	Valid bool `json:"valid"`
}

type cleanedResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.FileServerHits.Load())))
}

func (cfg *ApiConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {

	devMode := os.Getenv("PLATFORM")
	if devMode != "dev" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Printf("Action not allowed if not in dev mode")
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		cfg.FileServerHits.Store(0)
		cfg.DBQuery.ResetUsers(r.Context())
	}
}

func (cfg *ApiConfig) ValidateChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	w.Header().Set("Content-Type", "application/json")
	err := decoder.Decode(&params)
	if err != nil {
		cfg.respondWithErr(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if len(params.Body) >= 140 {
		cfg.respondWithErr(w, 400, "Chirp is too long")
	} else {
		cleaned := cfg.filterExpletives(params)
		cfg.respondWithJSON(w, http.StatusOK, cleaned)
	}
}

func (cfg *ApiConfig) AddUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	email := userReq{}
	w.Header().Set("Content-Type", "application/json")
	err := decoder.Decode(&email)
	if err != nil {
		cfg.respondWithErr(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.DBQuery.CreateUser(r.Context(), email.Email)
	if err != nil {
		cfg.respondWithErr(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	fmt.Println(user)

	customUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	cfg.respondWithJSON(w, 201, customUser)

}

func (cfg *ApiConfig) respondWithErr(w http.ResponseWriter, statusCode int, msg string) {
	errMsg := errorResponse{
		Error: msg,
	}
	cfg.respondWithJSON(w, statusCode, errMsg)
}

func (cfg *ApiConfig) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.WriteHeader(statusCode)
	dat, _ := json.Marshal(payload)
	w.Write(dat)
}

func (cfg *ApiConfig) filterExpletives(params parameters) cleanedResponse {
	expletives := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	splitBody := strings.Split(params.Body, " ")
	for i, words := range splitBody {
		if _, exists := expletives[strings.ToLower(words)]; exists {
			splitBody[i] = "****"
		}
	}
	cleaned := cleanedResponse{
		CleanedBody: strings.Join(splitBody, " "),
	}
	return cleaned
}
