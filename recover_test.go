package goutil

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	panicFn := func() { panic("Test panic for Go") }

	// Capture the panic output for the default recover function
	original := log.Writer()
	defer log.SetOutput(original)
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	Go(panicFn)

	// Allow some time for the goroutine to execute
	time.Sleep(500 * time.Millisecond)

	w.Close()
	out, _ := ioutil.ReadAll(r)

	if !strings.Contains(string(out), "Test panic for Go") {
		t.Errorf("Expected 'Test panic for Go', got '%v'", string(out))
	}
}

func TestGoWithErrorHandler(t *testing.T) {
	panicFn := func() { panic("Test panic for GoWithErrorHandler") }

	errorHandler := func(err interface{}) {
		if err != "Test panic for GoWithErrorHandler" {
			t.Errorf("Expected 'Test panic for GoWithErrorHandler', got '%v'", err)
		}
	}

	GoWithErrorHandler(panicFn, errorHandler)

	// Allow some time for the goroutine to execute
	time.Sleep(500 * time.Millisecond)
}
