package logger

import (
  corelog "log"
  "os"
)

var Verbose = false

func Log(errType string, text string) {
  if (errType == "error") || (errType == "fatal") || (Verbose == true) {
    corelog.Print("AHOY! [", errType, "] ==>", text, "\n")
  }
  if errType == "fatal" {
    os.Exit(1)
  }
}
