package util

import "log"

var Logging = false

func Logf(format string, args ...interface{}) {
	if !Logging {
		return
	}
	log.Printf(format, args...)
}
