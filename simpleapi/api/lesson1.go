// Package lesson1 provides primitives to interact with the openapi HTTP API.
//
// This module is hand-written, from the petstore example petstore.go, and implements the API.

package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Lession1 struct {
	Something int64
	Lock   sync.Mutex
}

// Make sure we conform to ServerInterface

var _ ServerInterface = (*Lession1)(nil)

func NewLession1() *Lession1 {
	return &Lession1{
		Something:   1,
	}
}

// sendLession1Error wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendLession1Error(w http.ResponseWriter, code int, message string) {
	lession1Err := Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(lession1Err)
}

// FindLesson1s implements all the handlers in the ServerInterface
func (p *Lession1) Gethw(w http.ResponseWriter, r *http.Request) {
	var result string

	result = "Hello World"

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
