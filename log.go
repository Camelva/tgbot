package main

import (
	"fmt"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"time"
)

var LogFile string

func initLog() *logrus.Logger {
	LogFile = fmt.Sprintf("%s.log", time.Now().Format("2006-01-02 15-04"))

	l := logrus.New()
	l.Hooks.Add(lfshook.NewHook(LogFile, &logrus.JSONFormatter{}))

	l.SetLevel(logrus.TraceLevel)

	return l
}
