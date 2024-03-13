package redis

import (
	"booking-service/config"

	"github.com/redis/go-redis/v9"
)

func SetupClient(cfg *config.RedisConfig) *redis.Client {
	conn := redis.NewClient(&redis.Options{
		Addr:            cfg.Host + ":" + cfg.Port,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      cfg.MaxRetries,
		PoolFIFO:        cfg.PoolFIFO,
		PoolSize:        cfg.PoolSize,
		PoolTimeout:     cfg.PoolTimeout,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	return conn
}
