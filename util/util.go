package util

import "log"

// Logging is a clumsy switch that affects what Logf does.
//
// If Logging is true, then Logf calls log.Printf.
var Logging = false

// Logf is a silly utility function that calls log.Printf if Logging
// is true.
func Logf(format string, args ...interface{}) {
	if !Logging {
		return
	}
	log.Printf(format, args...)
}
