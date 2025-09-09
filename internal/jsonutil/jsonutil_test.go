package jsonutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    Envelope
	}{
		{
			name:       "simple message",
			statusCode: 200,
			message:    Envelope{"message": "hello"},
		},
		{
			name:       "multiple fields",
			statusCode: 200,
			message:    Envelope{"id": 1.0, "status": "created"}, // id is 1.0 because Unmarshal
			// always decodes to float64, so we would then be comparing int(1) to float64(1) in our
			// tests which would fail
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			err := WriteJSON(rr, tc.statusCode, tc.message)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if rr.Code != tc.statusCode {
				t.Fatalf("expected code=%d, got code=%d", tc.statusCode, rr.Code)
			}

			var got Envelope
			if err = json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON written: %s", err)
			}

			for k, v := range tc.message {
				if got[k] != v {
					t.Fatalf("expected field %q=%v, got %v", k, v, got[k])
				}
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	type mockUser struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	tests := []struct {
		name           string
		body           string
		expectedErrMsg string
	}{
		{
			name: "Valid JSON",
			body: `{"id": 1, "name": "yusuf"}`,
		},
		{
			name:           "malformed JSON",
			body:           `{"id": 1, "name": "yusuf",}`, // trailing comma
			expectedErrMsg: "body contains badly formed JSON",
		},
		{
			name:           "wrong type",
			body:           `{"id": "abc", "name": "yusuf"}`,
			expectedErrMsg: "body contains incorrect type for field: id",
		},
		{
			name:           "unknown field",
			body:           `{"id": 1, "name": "yusuf", "extra": "oops"}`,
			expectedErrMsg: "body contains unknown key: \"extra\"",
		},
		{
			name:           "empty body",
			body:           ``,
			expectedErrMsg: "body cannot be empty",
		},
		{
			name:           "multiple JSON values",
			body:           `{"id": 1, "name": "yusuf"},{"id": 2, "name": "mohamed"}`,
			expectedErrMsg: "body must contain only one JSON value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()

			var dst mockUser

			err := ReadJSON(rr, req, &dst)

			if tc.expectedErrMsg == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Fatalf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			}
		})
	}
}
