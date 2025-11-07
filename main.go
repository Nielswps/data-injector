package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
	redis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

type Config struct {
	RedisAddress       string
	DataPath           string
	PingMaxAttempts    int
	PingAttemptTimeout time.Duration
}

type KeyValue struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

type RedisClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	JSONSet(ctx context.Context, key string, path string, value interface{}) *redis.StatusCmd
	Close() error
}

func main() {
	_ = godotenv.Load()

	var cfg Config

	rootCmd := &cobra.Command{
		Use:   "data-injector",
		Short: "Inject static JSON data into various storage solutions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fillMissingConfigFromEnv(&cfg)
			if cfg.RedisAddress == "" || cfg.DataPath == "" {
				return errors.New("redis address and data file path must be specified")
			}
			return run(context.Background(), cfg, nil)
		},
	}

	rootCmd.Flags().StringVarP(&cfg.RedisAddress, "redis-address", "r", "", "Address of Redis (e.g., localhost:6379)")
	rootCmd.Flags().StringVarP(&cfg.DataPath, "data-file", "f", "", "Path to JSON file containing data to inject")
	rootCmd.Flags().IntVar(&cfg.PingMaxAttempts, "ping-max-attempts", 10, "Max ping attempts")
	rootCmd.Flags().DurationVar(&cfg.PingAttemptTimeout, "ping-attempt-delay", 3*time.Second, "Delay between ping retries")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func fillMissingConfigFromEnv(cfg *Config) {
	if cfg.RedisAddress == "" {
		cfg.RedisAddress = os.Getenv("REDIS_ADDRESS")
	}
	if cfg.DataPath == "" {
		cfg.DataPath = os.Getenv("DATA_PATH")
	}
	if cfg.PingMaxAttempts == 0 {
		if v := os.Getenv("PING_MAX_ATTEMPTS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				cfg.PingMaxAttempts = n
			}
		} else {
			cfg.PingMaxAttempts = 10
		}
	}
	if cfg.PingAttemptTimeout == 0 {
		if v := os.Getenv("PING_ATTEMPT_TIMEOUT"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				cfg.PingAttemptTimeout = d
			}
		} else {
			cfg.PingAttemptTimeout = 3 * time.Second
		}
	}
}

func run(ctx context.Context, cfg Config, client RedisClient) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	if client == nil {
		client = redis.NewClient(&redis.Options{Addr: cfg.RedisAddress})
		defer client.Close()
	}

	log.Info().Str("redis_addr", cfg.RedisAddress).Msg("Connecting to Redis")
	if err := pingRedis(ctx, client, cfg.PingMaxAttempts, cfg.PingAttemptTimeout); err != nil {
		return err
	}
	log.Info().Msg("Connected to Redis")

	data, err := os.ReadFile(cfg.DataPath)
	if err != nil {
		log.Error().Err(err).Str("path", cfg.DataPath).Msg("Failed to read JSON file")
		return fmt.Errorf("could not read JSON file: %w", err)
	}

	var records []KeyValue
	if err := json.Unmarshal(data, &records); err != nil {
		log.Error().Err(err).Msg("Invalid JSON structure")
		return fmt.Errorf("could not unmarshal JSON: %w", err)
	}

	log.Info().Int("records", len(records)).Msg("Starting data injection")

	for _, kv := range records {
		if kv.Key == "" {
			log.Error().Interface("record", kv).Msg("Missing key in record")
			return fmt.Errorf("missing key in record: %v", kv)
		}

		if err := client.JSONSet(ctx, kv.Key, ".", kv.Value).Err(); err != nil {
			log.Error().Err(err).Str("key", kv.Key).Msg("Failed to write key")
			return fmt.Errorf("failed to set key '%s': %w", kv.Key, err)
		}

		log.Info().Str("key", kv.Key).Msg("Wrote JSON key")
	}

	log.Info().Msg("All data successfully written to Redis")
	return nil
}

func pingRedis(ctx context.Context, client RedisClient, maxAttempts int, retryDelay time.Duration) error {
	attempt := 1
	delay := retryDelay

	for {
		err := client.Ping(ctx).Err()
		if err == nil {
			log.Info().
				Int("attempt", attempt).
				Msg("Successfully connected to Redis")
			return nil
		}

		log.Warn().
			Err(err).
			Int("attempt", attempt).
			Msg("Failed to connect to Redis, retrying...")

		if attempt >= maxAttempts {
			return fmt.Errorf("unable to connect to Redis after %d attempts: %w", attempt, err)
		}

		time.Sleep(delay)
		delay *= 2 // exponential backoff
		attempt++
	}
}
