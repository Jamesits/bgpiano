package main

import (
	"github.com/osrg/gobgp/v3/pkg/log"
	"github.com/sirupsen/logrus"
)

// logrusLogger implements github.com/osrg/gobgp/v3/pkg/log/Logger interface
type logrusLogger struct {
	logger *logrus.Logger
}

func (l *logrusLogger) Panic(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Panic(msg)
}

func (l *logrusLogger) Fatal(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Fatal(msg)
}

func (l *logrusLogger) Error(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Error(msg)
}

func (l *logrusLogger) Warn(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Warn(msg)
}

func (l *logrusLogger) Info(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Info(msg)
}

func (l *logrusLogger) Debug(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Debug(msg)
}

func (l *logrusLogger) SetLevel(level log.LogLevel) {
	l.logger.SetLevel(logrus.Level(level))
}

func (l *logrusLogger) GetLevel() log.LogLevel {
	return log.LogLevel(l.logger.GetLevel())
}
