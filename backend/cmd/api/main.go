package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"presupuesto-rapido/backend/internal/config"
	"presupuesto-rapido/backend/internal/database"
	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/handlers"
	"presupuesto-rapido/backend/internal/httpx"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var db = mustConnectDatabase(ctx, cfg.DatabaseURL)
	if db != nil {
		defer db.Close()
	}

	h := handlers.Handler{DB: db, BossEmail: cfg.BossEmail}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("GET /api/me", devAuth(h.Me, domain.RoleBoss))
	mux.Handle("GET /api/prices", devAuth(h.ListPrices, domain.RoleEmployee))
	mux.Handle("POST /api/prices", devAuth(h.CreatePrice, domain.RoleBoss))
	mux.Handle("GET /api/documents", devAuth(h.ListDocuments, domain.RoleEmployee))
	mux.Handle("POST /api/documents", devAuth(h.CreateDocument, domain.RoleEmployee))

	var app http.Handler = mux
	app = httpx.SecurityHeaders(app)
	app = httpx.CORS(cfg.CORSAllowed)(app)
	app = httpx.RequestTimeout(15 * time.Second)(app)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s (%s)", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func mustConnectDatabase(ctx context.Context, databaseURL string) anyDB {
	if databaseURL == "" {
		return nil
	}
	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		log.Printf("database unavailable: %v", err)
		return nil
	}
	return db
}

type anyDB interface{ Close() }

func devAuth(next http.HandlerFunc, fallbackRole domain.Role) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Temporary bootstrap auth while real JWT login is wired.
		// Set X-Dev-User-ID, X-Dev-User-Email and X-Dev-Role from a trusted local client only.
		role := domain.Role(r.Header.Get("X-Dev-Role"))
		if role == "" {
			role = fallbackRole
		}
		user := domain.SessionUser{
			ID:    headerOr(r, "X-Dev-User-ID", "00000000-0000-0000-0000-000000000001"),
			Email: headerOr(r, "X-Dev-User-Email", "dev@example.local"),
			Role:  role,
		}
		next.ServeHTTP(w, r.WithContext(httpx.WithUser(r.Context(), user)))
	})
}

func headerOr(r *http.Request, key, fallback string) string {
	if v := r.Header.Get(key); v != "" {
		return v
	}
	return fallback
}
