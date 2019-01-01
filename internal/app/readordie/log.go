package readordie

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

// InitLogging initializes logging
func InitLogging() {
	var filename = "readordie.log"
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
}
