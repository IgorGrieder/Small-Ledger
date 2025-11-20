package repo

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
	"github.com/redis/go-redis/v9"
)

func SetupRedis(cfg *cfg.Config) *redis.Client {

	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.REDIS_ADDR, cfg.REDIS_PORT),
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	err := redis.Ping(context.Background()).Err()
	if err != nil {
		slog.Error("failed connecting into Redis")

		os.Exit(1)
	}

	return redis
}
