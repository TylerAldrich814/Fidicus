package config

import (
  "os"
  "strings"
  "github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger initializess Logrus based on 'LogLevel' extracted from AppConfig.
func InitLogger(){
  Logger = logrus.New()

  switch strings.ToLower(GetAppConfig().LogLevel) {
  case "trace":
    Logger.SetLevel(logrus.TraceLevel)
  case "debug":
    Logger.SetLevel(logrus.DebugLevel)
  case "info":
    Logger.SetLevel(logrus.InfoLevel)
  case "warn":
    Logger.SetLevel(logrus.WarnLevel)
  case "error":
    Logger.SetLevel(logrus.ErrorLevel)
  case "fatal": 
    Logger.SetLevel(logrus.FatalLevel)
  default: 
    Logger.SetLevel(logrus.InfoLevel)
  }

  if GetAppConfig().Environment == "production" {
    Logger.SetFormatter(&logrus.JSONFormatter{})
  } else {
    Logger.SetFormatter(&logrus.TextFormatter{
      FullTimestamp: true,
    })
  }

  Logger.SetOutput(os.Stdout)
}
