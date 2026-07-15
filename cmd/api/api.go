package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alimohammadi/golan-social.git/docs"
	"github.com/alimohammadi/golan-social.git/internal/auth"
	"github.com/alimohammadi/golan-social.git/internal/env"
	"github.com/alimohammadi/golan-social.git/internal/mailer"
	"github.com/alimohammadi/golan-social.git/internal/store"
	"github.com/alimohammadi/golan-social.git/internal/store/cache"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	db            dbConfing
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	cacheStorage  cache.Storage
}

type dbConfing struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr        string
	db          dbConfing
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redisCfg    redisConfig
}

type redisConfig struct {
	addr     string
	password string
	db       int
	enabled  bool
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

type mailConfig struct {
	exp       time.Duration
	fromEmail string
	sendGrid  sendGridConfig
}

type sendGridConfig struct {
	apiKey string
}

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()

	// A good middleware stack
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:5174")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route(
		"/v1", func(r chi.Router) {
			r.Get("/health", app.healthCheckHandler)

			docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
			r.Get("/swagger/*", httpSwagger.Handler(
				httpSwagger.URL(docsURL), //The url pointing to API definition
			))

			r.Route("/posts", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Post("/", app.createPostHandler)
				r.Route("/{postID}", func(r chi.Router) {
					r.Use(app.postsContextMiddleware)
					r.Get("/", app.getPostHandler)

					r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
					r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
				})
			})

			r.Route("/users", func(r chi.Router) {
				r.Put("/activate/{token}", app.activateUserHandler)

				r.Route("/{userID}", func(r chi.Router) {
					r.Use(app.AuthTokenMiddleware)

					r.Get("/", app.getUserHandler)
					r.Put("/follow", app.followUserHandler)
					r.Put("/unfollow", app.unfollowUserHandler)
				})

				r.Group(func(r chi.Router) {
					r.Get("/feed", app.getUserFeedHandler)
				})

			})

			// Public routes
			r.Route("/authentiction", func(r chi.Router) {
				r.Post("/user", app.registerUserHandler)
				// r.Post("/login", app.loginUserHandler)
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

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}
