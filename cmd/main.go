package main

import (
	"booking-service/config"
	"booking-service/internal/module/booking/handler"
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/database"
	"booking-service/internal/pkg/http"
	"booking-service/internal/pkg/httpclient"
	"booking-service/internal/pkg/log"
	"booking-service/internal/pkg/middleware"
	"booking-service/internal/pkg/redis"
	router "booking-service/internal/route"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.InitConfig()

	app := initService(cfg)

	// start http server
	http.StartHttpServer(app, cfg.HttpServer.Port)
}

func initService(cfg *config.Config) *fiber.App {

	// init database
	db := database.GetConnection(&cfg.Database)
	// init redis
	redis := redis.SetupClient(&cfg.Redis)
	// init logger
	logZap := log.SetupLogger()
	log.Init(logZap)
	logger := log.GetLogger()
	// init http client
	cb := httpclient.InitCircuitBreaker(&cfg.HttpClient, cfg.HttpClient.Type)
	httpClient := httpclient.InitHttpClient(&cfg.HttpClient, cb)

	// ctx := context.Background()
	// // init message stream
	// amqp := messagestream.NewAmpq(&cfg.MessageStream)

	// // Init Subscriber
	// subscriber, err := amqp.NewSubscriber()
	// if err != nil {
	// 	logger.Error(ctx, "Failed to create subscriber", err)
	// }

	// // Init Publisher
	// publisher, err := amqp.NewPublisher()
	// if err != nil {
	// 	logger.Error(ctx, "Failed to create publisher", err)
	// }

	// // Init message stream
	// go func() {
	// 	messages, err := subscriber.Subscribe(ctx, cfg.MessageStream.SubscribeTopic)
	// 	if err != nil {
	// 		logger.Error(ctx, "Failed to subscribe", err)
	// 	}
	// 	messagestream.ProcessMessages(messages)
	// }()

	ticketRepo := repositories.New(db, logger, httpClient, redis)
	// ticketUsecase := usecases.New(ticketRepo)
	middleware := middleware.Middleware{
		Repo: ticketRepo,
	}

	validator := validator.New()
	bookingHandler := handler.BookingHandler{
		Log:       logger,
		Validator: validator,
	}

	serverHttp := http.SetupHttpEngine()

	r := router.Initialize(serverHttp, &bookingHandler, &middleware)

	return r

}
