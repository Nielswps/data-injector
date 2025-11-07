package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	redis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

type Config struct {
	RedisConfigPath string
	DataPath        string
}

type KeyValue struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

func main() {
	var cfg Config

	var rootCmd = &cobra.Command{
		Use:   "data-injector",
		Short: "A tool to injecting static data into data storage.",
		Long: `A simplistic tool for injecting static data into different storage solutions, intended for local testing and showcasing.
Configurations can be provided via environment variables or command-line flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.RedisConfigPath == "" {
				cfg.RedisConfigPath = os.Getenv("REDIS_ADDRESS")
			}
			if cfg.DataPath == "" {
				cfg.DataPath = os.Getenv("DATA_PATH")
			}

			if cfg.RedisConfigPath == "" || cfg.DataPath == "" {
				fmt.Println("Error: Redis address and data file path must be specified.")
				os.Exit(1)
			}

			if err := run(cfg); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().StringVarP(&cfg.RedisConfigPath, "redis-address", "r", "", "Address that Redis is reachable at.")
	rootCmd.Flags().StringVarP(&cfg.DataPath, "data-file", "f", "", "Path to the data to upload (in JSON format).")

	godotenv.Load()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run(cfg Config) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisConfigPath,
	})

	ctx := context.Background()

	err := pingRedis(ctx, rdb)
	if err != nil {
		return fmt.Errorf("could not ping redis at: '%s' after retries. Failing", cfg.RedisConfigPath)
	}

	fmt.Println("Successfully connected to Redis!")

	data, err := os.ReadFile(cfg.DataPath)
	if err != nil {
		return fmt.Errorf("could not read JSON file: %w", err)
	}

	var jsonData []KeyValue
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	for _, item := range jsonData {
		if err := rdb.JSONSet(ctx, item.Key, ".", item.Value).Err(); err != nil {
			return fmt.Errorf("failed to set key '%s' in Redis: %w", item.Key, err)
		}
		fmt.Printf("Wrote key '%s' as JSON to Redis.\n", item.Key)
	}

	fmt.Println("Data successfully written to Redis as JSON.")
	return nil
}

func pingRedis(ctx context.Context, rdb *redis.Client) error {
	maxAttempts := 10
	attempt := 1
	for attempt <= maxAttempts {
		if err := rdb.Ping(ctx).Err(); err != nil {
			fmt.Println(fmt.Errorf("could not connect to Redis: %w", err))
			time.Sleep(3 * time.Second)
		} else {
			break
		}
		attempt += 1
	}

	if attempt > maxAttempts {
		return fmt.Errorf("unable to connecton to redis at %s", rdb.Options().Addr)
	}

	return nil
}
