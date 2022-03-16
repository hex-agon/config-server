package main

import (
	"context"
	"github.com/caarlos0/env/v6"
	"github.com/julienschmidt/httprouter"
	"github.com/runelite/config-server/runelite"
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
	Port            string `env:"PORT" envDefault:"8080"`
	MongodbUri      string `env:"MONGODB_URI,required,notEmpty"`
	MaxPayloadBytes int64  `env:"MAX_PAYLOAD_BYTES" envDefault:"5242880"` // 5mb default
}

type maxBytesHandler struct {
	handler  http.Handler
	maxBytes int64
}

func (h *maxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)
	h.handler.ServeHTTP(w, r)
}

func setupMongoDatabase(cfg config, logger *zap.Logger) (*mongo.Client, *mongo.Collection) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mongodb, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongodbUri))

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

func main() {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := loggerCfg.Build()

	defer logger.Sync()

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
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
	router := httprouter.New()
	handlers := runelite.NewHandlers(logger, runelite.NewRepository(cfgCollection))

	router.GET("/config", runelite.Authenticated(handlers.HandleGet))
	router.PUT("/config/:key", runelite.Authenticated(handlers.HandlePut))
	router.PATCH("/config/:key", runelite.Authenticated(handlers.HandlePatch))
	router.DELETE("/config/:key", runelite.Authenticated(handlers.HandleDelete))

	logger.Info("Starting server on port " + cfg.Port)
	err := http.ListenAndServe(":"+cfg.Port, &maxBytesHandler{handler: router, maxBytes: cfg.MaxPayloadBytes})
	if err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
