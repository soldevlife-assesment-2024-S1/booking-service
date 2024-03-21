package main

import (
	"booking-service/config"
	"booking-service/internal/module/booking/handler"
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/module/booking/usecases"
	"booking-service/internal/pkg/database"
	"booking-service/internal/pkg/http"
	"booking-service/internal/pkg/httpclient"
	log_internal "booking-service/internal/pkg/log"
	"booking-service/internal/pkg/messagestream"
	"booking-service/internal/pkg/middleware"
	"booking-service/internal/pkg/redis"
	router "booking-service/internal/route"
	"context"
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.InitConfig()

	app, messageRouters := initService(cfg)

	for _, router := range messageRouters {
		ctx := context.Background()
		go func(router *message.Router) {
			err := router.Run(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}(router)
	}

	// start http server
	http.StartHttpServer(app, cfg.HttpServer.Port)
}

func initService(cfg *config.Config) (*fiber.App, []*message.Router) {

	// init database
	db := database.GetConnection(&cfg.Database)
	// init redis
	redis := redis.SetupClient(&cfg.Redis)
	// init logger
	logZap := log_internal.SetupLogger()
	log_internal.Init(logZap)
	logger := log_internal.GetLogger()
	// init http client
	cb := httpclient.InitCircuitBreaker(&cfg.HttpClient, cfg.HttpClient.Type)
	httpClient := httpclient.InitHttpClient(&cfg.HttpClient, cb)

	ctx := context.Background()
	// init message stream
	amqp := messagestream.NewAmpq(&cfg.MessageStream)

	// Init Subscriber
	subscriber, err := amqp.NewSubscriber()
	if err != nil {
		logger.Error(ctx, "Failed to create subscriber", err)
	}

	// Init Publisher
	publisher, err := amqp.NewPublisher()
	if err != nil {
		logger.Error(ctx, "Failed to create publisher", err)
	}

	ticketRepo := repositories.New(db, logger, httpClient, redis)
	ticketUsecase := usecases.New(ticketRepo, logger, publisher)
	middleware := middleware.Middleware{
		Repo: ticketRepo,
	}

	validator := validator.New()
	bookingHandler := handler.BookingHandler{
		Log:       logger,
		Validator: validator,
		Usecase:   ticketUsecase,
		Publish:   publisher,
	}

	var messageRouters []*message.Router

	consumeBookingQueueRouter, err := messagestream.NewRouter(publisher, "book_ticket_poisoned", "book_ticket_handler", "book_ticket", subscriber, bookingHandler.ConsumeBookingQueue)
	if err != nil {
		logger.Error(ctx, "Failed to create consume_booking_queue router", err)
	}

	messageRouters = append(messageRouters, consumeBookingQueueRouter)

	serverHttp := http.SetupHttpEngine()

	r := router.Initialize(serverHttp, &bookingHandler, &middleware)

	return r, messageRouters

}
