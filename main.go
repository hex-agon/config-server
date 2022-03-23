package main

import (
	"context"
	"database/sql"
	"github.com/caarlos0/env/v6"
	_ "github.com/go-sql-driver/mysql"
	"github.com/newrelic/go-agent/v3/integrations/nrhttprouter"
	"github.com/newrelic/go-agent/v3/integrations/nrmongo"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

type config struct {
	Port                 string `env:"PORT" envDefault:"8080"`
	MongodbUri           string `env:"MONGODB_URI,required,notEmpty"`
	MysqlUri             string `env:"MYSQL_URI,required,notEmpty"`
	MysqlConnPool        int    `env:"MYSQL_POOL_SIZE" envDefault:"10"`
	MysqlConnLifetime    int    `env:"MYSQL_CONN_LIFETIME" envDefault:"5"`
	MaxPayloadBytes      int64  `env:"MAX_PAYLOAD_BYTES" envDefault:"5242880"` // 5mb default
	MaxConfigValueLength int64  `env:"MAX_CONFIG_VALUE_LENGTH" envDefault:"262144"`
	NewRelicLicense      string `env:"NR_LICENSE"`
}

type maxBytesHandler struct {
	handler  http.Handler
	maxBytes int64
}

func (h *maxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)
	h.handler.ServeHTTP(w, r)
}

func setupMongoDatabase(cfg *config, logger *zap.Logger) (*mongo.Client, *mongo.Collection) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	monitor := nrmongo.NewCommandMonitor(nil)
	mongodb, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongodbUri).SetMonitor(monitor))

	if err != nil {
		logger.Fatal("Failed to create mongodb client", zap.Error(err))
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// mongo.Connect doesn't actually connect, to verify if we can connect properly we have to Ping the server
	err = mongodb.Ping(ctx, readpref.Primary())

	if err != nil {
		logger.Fatal("Failed to ping mongodb", zap.Error(err))
	}
	cfgCollection := mongodb.Database("runelite").Collection("config")

	_, err = cfgCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"_userId": 1},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		logger.Fatal("Failed to create mongodb index", zap.Error(err))
	}
	return mongodb, cfgCollection
}

func setupMysql(cfg *config, logger *zap.Logger) *sql.DB {
	mysql, err := sql.Open("mysql", cfg.MysqlUri)
	if err != nil {
		logger.Fatal("Failed to connect to mysql", zap.Error(err))
	}
	err = mysql.Ping()
	if err != nil {
		logger.Fatal("Failed to ping mysql", zap.Error(err))
	}
	mysql.SetConnMaxLifetime(time.Minute * time.Duration(cfg.MysqlConnLifetime))
	mysql.SetMaxOpenConns(cfg.MysqlConnPool)
	mysql.SetMaxIdleConns(cfg.MysqlConnPool)
	return mysql
}

func main() {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := loggerCfg.Build()

	defer logger.Sync()

	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		logger.Fatal("Failed to load env config", zap.Error(err))
	}
	mongodb, cfgCollection := setupMongoDatabase(cfg, logger)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongodb.Disconnect(ctx); err != nil {
			logger.Fatal("Failed to disconnect from mongodb", zap.Error(err))
		}
	}()

	mysql := setupMysql(cfg, logger)
	defer mysql.Close()

	nrelic, err := newrelic.NewApplication(
		newrelic.ConfigAppName("config-server"),
		newrelic.ConfigLicense(cfg.NewRelicLicense),
		newrelic.ConfigEnabled(cfg.NewRelicLicense != ""),
	)

	if cfg.NewRelicLicense != "" {
		logger.Info("NewRelic agent is enabled")
	}
	router := nrhttprouter.New(nrelic)
	handlers := NewHandlers(logger, NewConfigRepository(cfgCollection, cfg.MaxConfigValueLength))

	sessionCache, err := NewSessionCache(NewSessionRepository(mysql), 10000)
	if err != nil {
		logger.Fatal("Failed to create session cache", zap.Error(err))
	}
	authFilter := NewAuthFilter(sessionCache)

	router.GET("/config", authFilter.Filtered(handlers.HandleGet))
	router.PUT("/config/:key", authFilter.Filtered(handlers.HandlePut))
	router.PATCH("/config", authFilter.Filtered(handlers.HandlePatch))
	router.DELETE("/config/:key", authFilter.Filtered(handlers.HandleDelete))

	logger.Info("Starting server on port " + cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, &maxBytesHandler{handler: router, maxBytes: cfg.MaxPayloadBytes})
	if err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
