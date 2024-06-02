package messagestream

import (
	"os"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
)

var (
	stateLog, _ = strconv.ParseBool(os.Getenv("PRODUCTION"))
)

type MessageStream interface {
	NewSubscriber() (message.Subscriber, error)
	NewPublisher() (message.Publisher, error)
}

func NewRouter(pub message.Publisher, poisonTopic string, handlerTopicName string, subscribeTopic string, subs message.Subscriber, handlerFunc func(msg *message.Message) error) (*message.Router, error) {
	logger := watermill.NewStdLogger(stateLog, stateLog)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddPlugin(plugin.SignalsHandler)

	poisonMiddleware, err := middleware.PoisonQueue(
		pub,
		poisonTopic,
	)

	if err != nil {
		return nil, err
	}

	router.AddMiddleware(
		middleware.CorrelationID,
		poisonMiddleware,

		middleware.Retry{
			MaxRetries:      5,
			InitialInterval: 500,
			Logger:          logger,
		}.Middleware,

		middleware.CorrelationID,
		middleware.Recoverer,
		wotelfloss.ExtractRemoteParentSpanContext(),
		wotel.Trace(),
	)

	router.AddNoPublisherHandler(
		handlerTopicName,
		subscribeTopic,
		subs,
		handlerFunc,
	)

	router.AddPlugin(plugin.SignalsHandler)

	return router, err
}
