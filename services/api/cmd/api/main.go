package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/icerde/api/graph"
	"github.com/icerde/api/internal/activate"
	"github.com/icerde/api/internal/auth"
	"github.com/icerde/api/internal/connect"
	"github.com/icerde/api/internal/factory"
	"github.com/icerde/api/internal/llm"
	"github.com/icerde/api/internal/maestro"
	"github.com/icerde/api/internal/mailer"
	"github.com/icerde/api/internal/opencode"
	"github.com/icerde/api/internal/sidecar"
	"github.com/icerde/api/internal/store"
	"github.com/rs/cors"
)

func main() {
	addr := getenv("ICERDE_API_ADDR", "127.0.0.1:43148")
	webOrigin := getenv("ICERDE_WEB_ORIGIN", "http://127.0.0.1:43147")
	mongoURI := os.Getenv("MONGO_URI")
	pepper := getenv("ICERDE_CODE_PEPPER", "icerde-dev-pepper-change-me")

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
			log.Printf("mongo ping ok; auth collections still memory until mongo adapters")
		}
	}

	mail := &mailer.Service{
		Store:      memory,
		WebURL:     webOrigin,
		ResendKey:  os.Getenv("RESEND_API_KEY"),
		ResendFrom: os.Getenv("RESEND_FROM"),
		Require:    envTruthy("ICERDE_MAIL_REQUIRE"),
		SMTP: mailer.Config{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     getenv("SMTP_PORT", "587"),
			User:     os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     getenv("SMTP_FROM", "İçerde <icerde@localhost>"),
		},
	}
	log.Printf("mailer channel=%s require=%v", mail.Channel(), mail.Require)
	authSvc := auth.New(memory, mail, pepper, webOrigin)
	if err := llm.Seed(context.Background(), memory); err != nil {
		log.Fatalf("llm seed: %v", err)
	}
	llmSvc := &llm.Service{
		Store:     memory,
		Completer: llm.NewCompleter(os.Getenv("ICERDE_LLM_API_KEY"), os.Getenv("ICERDE_LLM_BASE_URL"), os.Getenv("ICERDE_LLM_MODEL")),
	}
	projectsRoot := getenv("ICERDE_PROJECTS_ROOT", "")
	if projectsRoot == "" {
		projectsRoot = filepath.Join("..", "..", "var", "projects")
	}
	if abs, err := filepath.Abs(projectsRoot); err == nil {
		projectsRoot = abs
	}
	fact := factory.New(memory, projectsRoot)
	fact.LLM = llmSvc
	oc := opencode.NewCLI()
	fact.OpenCode = oc
	fact.Activator = activate.New()
	fact.MaestroRun = maestro.New()
	ocBin, ocVer, ocOK := oc.Probe()
	if ocOK {
		log.Printf("opencode bin=%s version=%s", ocBin, ocVer)
	} else {
		log.Printf("opencode missing — scaffold kept, no fake write (bundle in vendor/bin)")
	}
	if hit, err := sidecar.Look("maestro"); err == nil {
		log.Printf("maestro bin=%s source=%s", hit.Path, hit.Source)
	} else {
		log.Printf("maestro missing — flows SKIPPED without a device")
	}
	log.Printf("projects root=%s llm=%s", projectsRoot, llmSvc.Completer.Channel())
	resolver := &graph.Resolver{
		Auth:    authSvc,
		Factory: fact,
		LLM:     llmSvc,
		Connect: &connect.Service{Store: memory, Git: connect.CLIGit{}},
	}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	mux := http.NewServeMux()
	mux.Handle("/graphql", withRequestMeta(srv))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := "memory"
		if mongoOK {
			status = "memory+mongo-ping"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"store":    status,
			"version":  "0.6.0-local",
			"mail":     mail.Channel(),
			"gdpr":     true,
			"llm":      llmSvc.Completer.Channel(),
			"opencode": sidecarStatus("opencode"),
			"maestro":  sidecarStatus("maestro"),
		})
	})
	mux.HandleFunc("/export/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/export/"), "/")
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		user, _, err := authSvc.SessionUser(r.Context(), token)
		if err != nil {
			http.Error(w, "oturum gerekli", http.StatusUnauthorized)
			return
		}
		path, err := fact.ZipPath(r.Context(), user.ID, id)
		if err != nil {
			http.Error(w, "zip henüz yok", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="icerde.zip"`)
		http.ServeFile(w, r, path)
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

func withRequestMeta(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token = strings.TrimSpace(token)
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = strings.TrimSpace(strings.Split(fwd, ",")[0])
		}
		ctx := graph.WithToken(r.Context(), token)
		ctx = graph.WithIP(ctx, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sidecarStatus(name string) string {
	hit, err := sidecar.Look(name)
	if err != nil {
		return "missing"
	}
	return hit.Source
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envTruthy(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
