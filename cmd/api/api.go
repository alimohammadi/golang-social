package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alimohammadi/golan-social.git/docs"
	"github.com/alimohammadi/golan-social.git/internal/store"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
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
	addr   string
	db     dbConfing
	env    string
	apiURL string
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

			docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
			r.Get("/swagger/*", httpSwagger.Handler(
				httpSwagger.URL(docsURL), //The url pointing to API definition
			))

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

				r.Group(func(r chi.Router) {
					r.Get("/feed", app.getUserFeedHandler)
				})
			})
		})

	return r
}

func (app *application) run(mux *chi.Mux) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

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
