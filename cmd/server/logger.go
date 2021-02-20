package main

import "github.com/sirupsen/logrus"

func ServerEventLogger(args ...interface{}) {
	logrus.Debug(args...)
}

func RequestLogger(args ...interface{}) {
	logrus.Info(args...)
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}
