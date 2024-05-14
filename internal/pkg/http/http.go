package http

import (
	"booking-service/config"
	"log"
	"os"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupHttpEngine() *fiber.App {
	// init http server
	app := fiber.New(
		fiber.Config{
			JSONEncoder: json.Marshal,
			JSONDecoder: json.Unmarshal,
		})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: false,
	}),
	)

	// setup tracer
	app.Use(otelfiber.Middleware())

	// setup logger
	app.Use(logger.New(logger.Config{
		Next:         nil,
		Done:         nil,
		Format:       "[${time}] ${status} - ${latency} ${method} ${path} ${TagReqHeaders} ${body} ${error} \n",
		TimeFormat:   "15:04:05",
		TimeZone:     "Local",
		TimeInterval: 500 * time.Millisecond,
		Output:       os.Stdout,
	}))

	return app

}

func InitTracer(cfg *config.Config) *sdktrace.TracerProvider {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(cfg.ServiceName),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func StartHttpServer(app *fiber.App, port string) {
	log.Fatal(app.Listen(":" + port))
}
