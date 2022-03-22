package logging_config

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

func LoggerConfig(logger *logrus.Logger, debug bool) {
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: debug,
	})

	logger.SetOutput(colorable.NewColorableStderr())

	if debug {
		logger.SetLevel(logrus.TraceLevel)
		logger.SetReportCaller(true)
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.SetReportCaller(false)
	}
}
