package main

import (
	"time"

	"github.com/alimohammadi/golan-social.git/internal/db"
	"github.com/alimohammadi/golan-social.git/internal/env"
	"github.com/alimohammadi/golan-social.git/internal/mailer"
	"github.com/alimohammadi/golan-social.git/internal/store"
	"github.com/alimohammadi/golan-social.git/internal/store/cache"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GopherSocial API
//	@description	API for GopherSocial, a social network for gopher
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1
//

//	@securityDefinitions.apiKey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {
	cfg := config{
		addr:   env.GetString("ADDR", ":8084"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8084"),
		db: dbConfing{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_Time", "15m"),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3 days
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		redisCfg: redisConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PW", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", false),
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "gophersocial",
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Connect to database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	// Cache
	var rdb *redis.Client
	cacheStorage := cache.NewRedisStorage(rdb)

	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(
			cfg.redisCfg.addr,
			cfg.redisCfg.password,
			cfg.redisCfg.db,
		)

		logger.Info("redis client established")
	}

	// Store
	store := store.NewStorage(db)

	mailer := mailer.NewSendgrid(
		cfg.mail.sendGrid.apiKey,
		cfg.mail.fromEmail,
	)

	app := &application{
		config:       cfg,
		store:        store,
		logger:       logger,
		mailer:       mailer,
		cacheStorage: cacheStorage,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
