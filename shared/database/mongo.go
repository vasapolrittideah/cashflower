package database

import (
	"context"
	"errors"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const defaultConnectionTimeout = 20 * time.Second

// MongoDB represents a MongoDB database connection.
type MongoDB struct {
	config   *mongoConfig
	client   *mongo.Client
	database *mongo.Database
	logger   *zerolog.Logger
}

// NewMongoDB creates a new MongoDB instance.
func NewMongoDB(logger *zerolog.Logger) *MongoDB {
	cfg := newMongoConfig(logger)

	if err := cfg.validate(); err != nil {
		logger.Fatal().Err(err).Msg("failed to validate MongoDB configuration")
	}

	return &MongoDB{
		config: cfg,
		logger: logger,
	}
}

// Connect establishes a connection to MongoDB.
func (d *MongoDB) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(d.config.URI))
	if err != nil {
		return err
	}

	d.client = client
	d.database = client.Database(d.config.DB)

	if err := d.client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	d.logger.Info().Str("uri", d.config.URI).Msg("Successfully connected to MongoDB")

	return nil
}

// Disconnect closes the MongoDB connection.
func (d *MongoDB) Disconnect(ctx context.Context) error {
	if err := d.client.Disconnect(ctx); err != nil {
		return err
	}

	d.logger.Info().Msg("Successfully disconnected from MongoDB")

	return nil
}

// GetDatabase returns the MongoDB database.
func (d *MongoDB) GetDatabase() *mongo.Database {
	return d.database
}

// MongoConfig contains MongoDB connection configuration.
type mongoConfig struct {
	URI string `env:"MONGO_URI"`
	DB  string `env:"MONGO_DB"`
}

// newMongoConfig creates a new MongoConfig instance from environment variables.
func newMongoConfig(logger *zerolog.Logger) *mongoConfig {
	cfg, err := env.ParseAs[mongoConfig]()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse environment variables")
	}

	return &cfg
}

// validate checks if the MongoDB configuration is valid.
func (c *mongoConfig) validate() error {
	if c.URI == "" {
		return errors.New("missing MONGO_URI environment variable")
	}

	if c.DB == "" {
		return errors.New("missing MONGO_DB environment variable")
	}

	return nil
}
