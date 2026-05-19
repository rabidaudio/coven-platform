package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

// UnmarshalBody takes an HTTP request body and deserializes the JSON into out.
// If deserialization fails, it will automatically render the error and return false.
func UnmarshalBody[T any](w http.ResponseWriter, r *http.Request, out *T) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		RespondBadFormat(w, fmt.Errorf("invalid request body: %w", err))
		return false
	}
	return true
}

// MarshalBody serializes in into JSON and writes it to the response body.
// If serialization fails, it will panic.
func MarshalBody[T any](w http.ResponseWriter, in *T) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(in); err != nil {
		panic(fmt.Errorf("serialize %v: %w", reflect.TypeOf(in).Name(), err))
	}
}
