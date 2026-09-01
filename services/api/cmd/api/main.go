package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/icerde/api/graph"
	"github.com/icerde/api/internal/store"
	"github.com/rs/cors"
)

func main() {
	addr := getenv("ICERDE_API_ADDR", "127.0.0.1:43148")
	webOrigin := getenv("ICERDE_WEB_ORIGIN", "http://127.0.0.1:43147")
	mongoURI := os.Getenv("MONGO_URI")

	memory := store.NewMemory()
	var mongoOK bool
	if mongoURI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		client, err := store.TryMongo(ctx, mongoURI)
		cancel()
		if err != nil {
			log.Printf("mongo unavailable, using memory store: %v", err)
		} else {
			mongoOK = true
			defer func() { _ = client.Disconnect(context.Background()) }()
			log.Printf("mongo ping ok; session store is still memory until auth slice")
		}
	}

	resolver := &graph.Resolver{Store: memory}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	mux := http.NewServeMux()
	mux.Handle("/graphql", withToken(srv))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := "memory"
		if mongoOK {
			status = "memory+mongo-ping"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"store":   status,
			"version": "0.1.0-scaffold",
		})
	})

	handlerWithCORS := cors.New(cors.Options{
		AllowedOrigins:   []string{webOrigin, "http://localhost:43147", "http://127.0.0.1:43147"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           handlerWithCORS,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("icerde api listening on http://%s/graphql", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func withToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token = strings.TrimSpace(token)
		ctx := graph.WithToken(r.Context(), token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
