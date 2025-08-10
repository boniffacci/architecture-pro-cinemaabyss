package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	port := getEnv("PORT", "8000")

	http.HandleFunc("/api/movies", moviesHandler)
	http.HandleFunc("/api/events", proxyHandler(getEnv("EVENTS_SERVICE_URL", "http://events-service:8082")))
	http.HandleFunc("/", proxyHandler(getEnv("MONOLITH_URL", "http://monolith:8080")))

	log.Printf("Proxy service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func moviesHandler(w http.ResponseWriter, r *http.Request) {
	gradualMigration := getEnv("GRADUAL_MIGRATION", "false") == "true"
	migrationPercent, _ := strconv.Atoi(getEnv("MOVIES_MIGRATION_PERCENT", "0"))

	moviesServiceURL := getEnv("MOVIES_SERVICE_URL", "http://movies-service:8081")
	monolithURL := getEnv("MONOLITH_URL", "http://monolith:8080")

	if gradualMigration {
		rand.Seed(time.Now().UnixNano())
		if rand.Intn(100) < migrationPercent {
			log.Println("Routing to movies microservice")
			proxyHandler(moviesServiceURL)(w, r)
			return
		}
	}

	log.Println("Routing to monolith")
	proxyHandler(monolithURL)(w, r)
}

func proxyHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := url.Parse(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
