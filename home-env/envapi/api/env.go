// Package lesson1 provides primitives to interact with the openapi HTTP API.
//
// This module is hand-written, from the petstore example petstore.go, and implements the API.

package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Env1 struct {
	Something int64
	Lock   sync.Mutex
}

// Make sure we conform to ServerInterface

var _ ServerInterface = (*Env1)(nil)

func NewEnv1() *Env1 {
	return &Env1{
		Something:   1,
	}
}

// sendEnv1 wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendEnv1(w http.ResponseWriter, code int, message string) {
	lession1Err := Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(lession1Err)
}

// FindEnv1s implements all the handlers in the ServerInterface
func (p *Env1) GetEnv(w http.ResponseWriter, r *http.Request) {
	var result GetEnv200JSONResponse 

	result.Temp = &temp
	result.Lux = &lux

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
