package log

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

func Setup() *otelzap.Logger {
	logger := otelzap.New(zap.NewExample())
	defer logger.Sync()

	undo := otelzap.ReplaceGlobals(logger)
	defer undo()

	return logger
}
