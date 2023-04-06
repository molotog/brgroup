package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Info(string, ...logrus.Fields)
	Error(string, error, ...logrus.Fields)
}

func NewLogger() Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return logger{
		log: log,
	}
}

type logger struct {
	log *logrus.Logger
}

func (l logger) Info(msg string, fields ...logrus.Fields) {
	logFields := make(logrus.Fields)
	for i := range fields {
		for key, value := range fields[i] {
			logFields[key] = value
		}
	}
	l.log.WithFields(logFields).Info(msg)
}

func (l logger) Error(msg string, err error, fields ...logrus.Fields) {
	logFields := make(logrus.Fields)
	for i := range fields {
		for key, value := range fields[i] {
			logFields[key] = value
		}
	}
	logFields["error"] = err
	l.log.WithFields(logFields).Error(msg)
}
