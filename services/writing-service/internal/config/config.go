package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort string
	
	// PostgreSQL
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	
	// MongoDB
	MongoHost     string
	MongoPort     int
	MongoUser     string
	MongoPassword string
	MongoDB       string
	
	// JWT
	JWTSecret string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8084"),
		
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "writing_db"),
		
		MongoHost:     getEnv("MONGO_HOST", "localhost"),
		MongoPort:     getEnvAsInt("MONGO_PORT", 27017),
		MongoUser:     getEnv("MONGO_USER", "admin"),
		MongoPassword: getEnv("MONGO_PASSWORD", "password"),
		MongoDB:       getEnv("MONGO_DB", "writing_doc_db"),
		
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
	}
}

// PostgresURL returns the PostgreSQL connection URL
func (c *Config) PostgresURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

// MongoURL returns the MongoDB connection URL
func (c *Config) MongoURL() string {
	if c.MongoUser != "" && c.MongoPassword != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d",
			c.MongoUser, c.MongoPassword, c.MongoHost, c.MongoPort)
	}
	return fmt.Sprintf("mongodb://%s:%d", c.MongoHost, c.MongoPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
