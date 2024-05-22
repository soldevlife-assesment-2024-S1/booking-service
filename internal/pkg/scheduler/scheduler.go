package scheduler

import (
	"booking-service/config"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

const (
	TypeSetPaymentExpired = "set_payment_expired"
)

type Scheduler struct {
	Log *otelzap.Logger
}

func (s *Scheduler) StartMonitoring(cfg *config.RedisConfig) {
	ctx := context.Background()
	redisAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	h := asynqmon.New(asynqmon.Options{
		RootPath: "/monitoring", // RootPath specifies the root for asynqmon app

		RedisConnOpt: asynq.RedisClientOpt{Addr: redisAddr, Password: cfg.Password, DB: cfg.DB},
	})

	// Note: We need the tailing slash when using net/http.ServeMux.
	http.Handle(h.RootPath()+"/", h)

	// Go to http://localhost:8080/monitoring to see asynqmon homepage.
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		s.Log.Ctx(ctx).Error(fmt.Sprintf("error start monitoring: %v", err))
	}
}

func (s *Scheduler) InitClient(cfg *config.RedisConfig) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})
}

func (s *Scheduler) StartHandler(cfg *config.RedisConfig, taskTypes []string, handlerFunc []func(ctx context.Context, t *asynq.Task) error) {
	ctx := context.Background()
	redisAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr, Password: cfg.Password, DB: cfg.DB},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)
	mux := asynq.NewServeMux()

	for i, taskType := range taskTypes {
		mux = s.registerHandlers(mux, taskType, handlerFunc[i])
	}

	if err := srv.Run(mux); err != nil {
		s.Log.Ctx(ctx).Error(fmt.Sprintf("error start handler: %v", err))
	}
}

func (s *Scheduler) registerHandlers(mux *asynq.ServeMux, typeTask string, handlerFunc func(ctx context.Context, t *asynq.Task) error) *asynq.ServeMux {
	// mux maps a type to a handler
	mux.HandleFunc(typeTask, handlerFunc)
	return mux
}
