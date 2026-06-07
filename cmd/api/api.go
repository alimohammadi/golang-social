package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alimohammadi/golan-social.git/internal/store"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type application struct {
	config config
	store  store.Storage
	db     dbConfing
}

type dbConfing struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr string
	db   dbConfing
	env  string
}

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()

	// A good middleware stack
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use()

	r.Route(
		"/v1", func(r chi.Router) {
			r.Get("/health", app.healthCheckHandler)

			r.Route("/posts", func(r chi.Router) {
				r.Post("/", app.createPostHandler)
				r.Route("/{postID}", func(r chi.Router) {
					r.Use(app.postsContextMiddleware)

					r.Get("/", app.getPostHandler)
					r.Patch("/", app.updatePostHandler)
					r.Delete("/", app.deletePostHandler)
				})
			})

			r.Route("/users", func(r chi.Router) {
				r.Route("/{userID}", func(r chi.Router) {
					r.Use(app.userContextMiddleware)
					
					r.Get("/", app.getUserHandler)
					r.Put("/follow", app.followUserHandler)
					r.Put("/unfollow", app.unfollowUserHandler)
				})
			})
		})
	// users
	// auth
	return r
}

func (app *application) run(mux *chi.Mux) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Print("server is started at $s", app.config.addr)

	return srv.ListenAndServe()
}
