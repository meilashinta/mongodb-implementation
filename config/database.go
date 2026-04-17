package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Database holds pooled connections for MySQL and MongoDB.
type Database struct {
	SQL               *sql.DB
	MongoClient       *mongo.Client
	ProductCollection *mongo.Collection
}

// InitDatabase creates a reusable database connection pool for MySQL and MongoDB.
func InitDatabase(ctx context.Context) (*Database, error) {
	mysqlDSN := getEnv("MYSQL_DSN", "appuser:password@tcp(localhost:3306)/dualwrite?parseTime=true")
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	mongoDatabase := getEnv("MONGODB_DATABASE", "dualwrite")
	mongoCollection := getEnv("MONGODB_COLLECTION", "products")

	// Initialize MySQL connection pool.
	sqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB.SetMaxOpenConns(getEnvAsInt("MYSQL_MAX_OPEN_CONNS", 25))
	sqlDB.SetMaxIdleConns(getEnvAsInt("MYSQL_MAX_IDLE_CONNS", 10))
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	clientOptions.SetMaxPoolSize(uint64(getEnvAsInt("MONGODB_MAX_POOL_SIZE", 100)))

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}

	collection := mongoClient.Database(mongoDatabase).Collection(mongoCollection)
	log.Printf("connected to MySQL and MongoDB; using collection %s.%s", mongoDatabase, mongoCollection)

	return &Database{SQL: sqlDB, MongoClient: mongoClient, ProductCollection: collection}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
