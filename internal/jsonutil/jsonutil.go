package jsonutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Envelope is a custom type for how the messages will be written to response
type Envelope map[string]any

// WriteJSON is a helper function to write responses
func WriteJSON(w http.ResponseWriter, statusCode int, message Envelope) error {
	// use MarshalIndent to make it indented and nice on terminal instead of flat text
	JSON, err := json.MarshalIndent(message, "", "\t")
	if err != nil {
		return err
	}

	// append new line to make it nice on terminal
	JSON = append(JSON, '\n')

	w.WriteHeader(statusCode)

	_, err = w.Write(JSON)
	return err
}

// ReadJSON is a helper function to read data from the request body to dst and returns error if
// something goes wrong
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// set a limit on the size of the request body so that very large requests don't hog resources
	const maxBytes = 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	// create a decoder and call the DisallowUnknownFields on it so that fields that are not in the
	// dst raise an error
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	// start decoding to the dst
	err := decoder.Decode(dst)
	if err != nil {
		// declare different types of errrors to hold them approprietly
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("body contains badly formed JSON at character: %d", syntaxErr.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly formed JSON")

		// this is when different variable type is passed to a field, like field expecting int and
		// passed string
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return fmt.Errorf(
					"body contains incorrect type for field: %s", unmarshalTypeErr.Field,
				)
			}
			return fmt.Errorf(
				"body contains incorrect type at character: %d", unmarshalTypeErr.Offset,
			)

		// this is when a key is in the request body that is not in the dst, the error is the
		// format "json: unknown field <name>"
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key: %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body size cannot exceed %d bytes", maxBytes)

		case errors.Is(err, io.EOF):
			return fmt.Errorf("body cannot be empty")

		// this is when an error that is not caused by the client occurs. this should not happen in
		// normal operations
		case errors.As(err, &invalidUnmarshalErr):
			panic(err)

		// for unknown errors
		default:
			return err
		}
	}

	// call the decode on an empty anonymous struct pointer to check if the request body is empty
	// if this doesn't return io.EOF it means the request body had more than one json value
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("body must contain only one JSON value")
	}

	return nil
}
