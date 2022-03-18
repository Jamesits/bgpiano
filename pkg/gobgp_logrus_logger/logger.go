package gobgp_logrus_logger

import (
	"github.com/osrg/gobgp/v3/pkg/log"
	"github.com/sirupsen/logrus"
)

// GobgpLogrusLogger implements github.com/osrg/gobgp/v3/pkg/log/Logger interface
type GobgpLogrusLogger struct {
	Logger *logrus.Logger
}

func (l *GobgpLogrusLogger) Panic(msg string, fields log.Fields) {
	l.Logger.WithFields(logrus.Fields(fields)).Panic(msg)
}

func (l *GobgpLogrusLogger) Fatal(msg string, fields log.Fields) {
	l.Logger.WithFields(logrus.Fields(fields)).Fatal(msg)
}

func (l *GobgpLogrusLogger) Error(msg string, fields log.Fields) {
	l.Logger.WithFields(logrus.Fields(fields)).Error(msg)
}

func (l *GobgpLogrusLogger) Warn(msg string, fields log.Fields) {
	l.Logger.WithFields(logrus.Fields(fields)).Warn(msg)
}

func (l *GobgpLogrusLogger) Info(msg string, fields log.Fields) {
	l.Logger.WithFields(logrus.Fields(fields)).Info(msg)
}

func (l *GobgpLogrusLogger) Debug(msg string, fields log.Fields) {
	l.Logger.WithFields(logrus.Fields(fields)).Debug(msg)
}

func (l *GobgpLogrusLogger) SetLevel(level log.LogLevel) {
	l.Logger.SetLevel(logrus.Level(level))
}

func (l *GobgpLogrusLogger) GetLevel() log.LogLevel {
	return log.LogLevel(l.Logger.GetLevel())
}
