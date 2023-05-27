package goutil

import (
	"log"
	"runtime/debug"
)

// Go starts a goroutine with recovery capability.
// If the goroutine panics, it will recover and use a default error handler.
func Go(fn func()) {
	GoWithErrorHandler(fn, defaultErrorHandler)
}

// GoWithErrorHandler starts a goroutine with a custom error handler.
// If the goroutine panics, it will recover and use the provided error handler.
func GoWithErrorHandler(fn func(), errorHandler func(err interface{})) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				errorHandler(err)
			}
		}()

		fn()
	}()
}

// defaultErrorHandler is a recovery function that logs the panic and prints the stack trace.
func defaultErrorHandler(err interface{}) {
	log.Println(err)
	debug.PrintStack()
}
