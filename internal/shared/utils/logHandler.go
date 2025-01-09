package utils

import (
  "fmt"
	log "github.com/sirupsen/logrus"
)

type LogLevel string
var (
  LogInfo LogLevel = "info"
  LogWarn LogLevel = "warn"
  LogErro LogLevel = "error"
  LogPani LogLevel = "panic"
)

// NewLogHandlerFunc -- Creates a custom wrapper around logrus.Error -- 
// returs a func that takes in the same paramters as fmt.Sprintf. Creating
// a formatted string before passing it to logrus.Error. Used for listing
// the provided function name, important Fields, and error information.
func NewLogHandlerFunc(
  fnName string, 
  fields log.Fields,
) func(LogLevel, string, ...any){
  return func(lvl LogLevel, f string, args ...any){
    switch lvl {
    case LogInfo:
      log.WithFields(fields).Info(fmt.Sprintf(fnName+": "+f, args...))
    case LogWarn:
      log.WithFields(fields).Warn(fmt.Sprintf(fnName+f, args...))
    case LogErro:
      log.WithFields(fields).Error(fmt.Sprintf(fnName+f, args...))
    default:
      log.WithFields(fields).Panic(fmt.Sprintf(fnName+f, args...))
    }
  }
}
