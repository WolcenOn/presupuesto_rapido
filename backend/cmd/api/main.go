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

	"github.com/jackc/pgx/v5/pgxpool"

	"presupuesto-rapido/backend/internal/config"
	"presupuesto-rapido/backend/internal/database"
	"presupuesto-rapido/backend/internal/handlers"
	"presupuesto-rapido/backend/internal/httpx"
	"presupuesto-rapido/backend/internal/mail"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := connectDatabase(ctx, cfg.DatabaseURL)
	if db != nil {
		defer db.Close()
	}
	if cfg.MailWorkerEnabled {
		mail.Worker{DB: db, Config: mail.Config{Host: cfg.SMTPHost, Port: cfg.SMTPPort, Username: cfg.SMTPUsername, Password: cfg.SMTPPassword, FromEmail: cfg.SMTPFromEmail, FromName: cfg.SMTPFromName}}.Start(ctx)
	}

	h := handlers.Handler{DB: db, BossEmail: cfg.BossEmail}
	authCfg := handlers.AuthConfig{JWTSecret: cfg.JWTSecret, BootstrapSecret: cfg.BootstrapSecret, AccessTokenTTL: cfg.AccessTokenTTL, RefreshTokenTTL: cfg.RefreshTokenTTL, CookieSecure: cfg.Env == "production"}
	authn := httpx.Authenticate(cfg.JWTSecret)
	bossOnly := chain(authn, httpx.RequireRole("boss"))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/setup/boss", h.SetupBoss(authCfg))
	mux.HandleFunc("POST /api/auth/login", h.Login(authCfg))
	mux.HandleFunc("POST /api/auth/refresh", h.Refresh(authCfg))
	mux.HandleFunc("POST /api/auth/logout", h.Logout(authCfg))
	mux.Handle("GET /api/me", authn(http.HandlerFunc(h.Me)))
	mux.Handle("GET /api/company", authn(http.HandlerFunc(h.GetCompany)))
	mux.Handle("PATCH /api/company", bossOnly(http.HandlerFunc(h.UpdateCompany)))
	mux.Handle("GET /api/prices", authn(http.HandlerFunc(h.ListPrices)))
	mux.Handle("POST /api/prices", bossOnly(http.HandlerFunc(h.CreatePrice)))
	mux.Handle("GET /api/prices/{id}", authn(http.HandlerFunc(h.GetPrice)))
	mux.Handle("PATCH /api/prices/{id}", bossOnly(http.HandlerFunc(h.UpdatePrice)))
	mux.Handle("DELETE /api/prices/{id}", bossOnly(http.HandlerFunc(h.DisablePrice)))
	mux.Handle("GET /api/documents", authn(http.HandlerFunc(h.ListDocuments)))
	mux.Handle("POST /api/documents", authn(http.HandlerFunc(h.CreateDocument)))
	mux.Handle("GET /api/documents/{id}", authn(http.HandlerFunc(h.GetDocument)))
	mux.Handle("POST /api/documents/{id}/send-to-boss", authn(http.HandlerFunc(h.QueueDocumentForBoss)))
	mux.Handle("GET /api/admin/users", bossOnly(http.HandlerFunc(h.ListUsers)))
	mux.Handle("POST /api/admin/users", bossOnly(http.HandlerFunc(h.CreateUser)))

	var app http.Handler = mux
	app = httpx.SecurityHeaders(app)
	app = httpx.CORS(cfg.CORSAllowed)(app)
	app = httpx.RequestTimeout(15 * time.Second)(app)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: app, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("api listening on :%s (%s)", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { log.Fatalf("server error: %v", err) }
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil { log.Printf("graceful shutdown failed: %v", err) }
}

func connectDatabase(ctx context.Context, databaseURL string) *pgxpool.Pool {
	if databaseURL == "" { return nil }
	db, err := database.Connect(ctx, databaseURL)
	if err != nil { log.Printf("database unavailable: %v", err); return nil }
	return db
}

func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- { final = middlewares[i](final) }
		return final
	}
}
