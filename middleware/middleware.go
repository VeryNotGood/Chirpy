package middleware

import (
	"VeryNotGood/Chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	DBQuery        *database.Queries
}

type parameters struct {
	Body string `json:"body"`
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
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.FileServerHits.Load())))
}

func (cfg *ApiConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.FileServerHits.Store(0)
}

func (cfg *ApiConfig) ValidateChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	w.Header().Set("Content-Type", "application/json")
	err := decoder.Decode(&params)
	if err != nil {
		cfg.respondWithErr(w, 500, "Something went wrong")
		return
	}
	if len(params.Body) >= 140 {
		cfg.respondWithErr(w, 400, "Chirp is too long")
	} else {
		cleaned := cfg.filterExpletives(w, params, r)
		cfg.respondWithJSON(w, 200, cleaned)
	}
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

func (cfg *ApiConfig) filterExpletives(w http.ResponseWriter, params parameters, r *http.Request) cleanedResponse {
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
