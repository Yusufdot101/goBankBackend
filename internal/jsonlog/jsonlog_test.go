package jsonlog

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name        string
		level       Level
		expectedStr string
	}{
		{
			name:        "level info",
			level:       LevelInfo,
			expectedStr: "INFO",
		},
		{
			name:        "level error",
			level:       LevelError,
			expectedStr: "ERROR",
		},
		{
			name:        "level fatal",
			level:       LevelFatal,
			expectedStr: "FATAL",
		},
		{
			name:        "level off",
			level:       LevelOff,
			expectedStr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStr := tc.level.ToString()
			if gotStr != tc.expectedStr {
				t.Fatalf("expected string %s, got %s", tc.expectedStr, gotStr)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name       string
		minLevel   Level
		level      Level
		message    string
		properties map[string]string
		wantLog    bool
		wantTrace  bool
	}{
		{
			name:     "below min level ignored",
			minLevel: LevelError,
			level:    LevelInfo,
			message:  "ignored message",
		},
		{
			name:     "info with properties",
			minLevel: LevelInfo,
			level:    LevelInfo,
			message:  "info msg",
			properties: map[string]string{
				"name":      "yusuf",
				"last name": "mohamed",
			},
			wantLog: true,
		},
		{
			name:      "error message includes trace",
			minLevel:  LevelInfo,
			level:     LevelError,
			message:   "error msg",
			wantLog:   true,
			wantTrace: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			l := New(buf, tc.minLevel)
			_, err := l.print(tc.level, tc.message, tc.properties)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !tc.wantLog {
				if buf.Len() != 0 {
					t.Fatalf("expected no log output, got %s", buf.String())
				}
				return
			}

			var entry struct {
				Level      string            `json:"level"`
				Message    string            `json:"message"`
				Properties map[string]string `json:"properties"`
				Trace      string            `json:"trace"`
				Time       string            `json:"time"`
			}

			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("invalid json: %v\n%s", err, buf.String())
			}

			if entry.Message != tc.message {
				t.Errorf("expected message %q, got %q", tc.message, entry.Message)
			}
			if tc.properties != nil && entry.Properties["name"] != "yusuf" {
				t.Errorf("expected property name=yusuf, got %v", entry.Properties)
			}
			if tc.wantTrace && entry.Trace == "" {
				t.Error("expected trace to be present, got empty string")
			}
			if !tc.wantTrace && entry.Trace != "" {
				t.Error("expected no trace, got:", entry.Trace)
			}
			if entry.Time == "" {
				t.Error("expected time to be set")
			}
		})
	}
}
