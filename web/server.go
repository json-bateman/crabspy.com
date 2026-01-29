package web

import (
	"context"
	"crabspy"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/go-chi/chi/v5"
)

//go:embed static/*
var StaticFS embed.FS

var (
	StaticSys = hashfs.NewFS(StaticFS)
)

func StaticPath(format string, args ...any) string {
	return "/" + StaticSys.HashName(fmt.Sprintf("static/"+format, args...))
}

func setupRoutes() chi.Router {
	r := chi.NewRouter()

	r.Handle("/static/*", hashfs.FileServer(StaticSys))
	r.Get("/", home())
	r.Get("/host", host())
	r.Get("/join", join())

	return r
}

func home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Home().Render(r.Context(), w)
	}
}

func host() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Host().Render(r.Context(), w)
	}
}
func join() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Join().Render(r.Context(), w)
	}
}

// RunBlocking sets up routes, starts the server, handles cleanup
func RunBlocking(setupCtx context.Context, db *sql.DB) error {
	router := setupRoutes()

	addr := fmt.Sprintf(":%d", crabspy.Env.Port)
	srv := http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		<-setupCtx.Done()
		log.Printf("I shutdown lmao")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Printf("Starting server on http://localhost%s", addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Error starting server: %v", err)
	}
	return nil
}
