package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/go-mail/mail/v2"
)

// fake dialer recipient capture sent emails
type fakeDialer struct {
	sent []*mail.Message
	fail bool
}

func (f *fakeDialer) DialAndSend(msg ...*mail.Message) error {
	if f.fail {
		return errors.New("dialer failed")
	}
	f.sent = append(f.sent, msg...)
	return nil
}

// --- actual test ---
func TestMailerSend(t *testing.T) {
	tests := []struct {
		name            string
		setupFakeDialer func(*fakeDialer)
		templateFile    string
		recipient       string
		data            map[string]any
		wantErr         bool
	}{
		{
			name: "valid",
			setupFakeDialer: func(f *fakeDialer) {
				f.sent = []*mail.Message{}
			},
			templateFile: "user_welcome.html",
			recipient:    "yusuf",
			data:         map[string]any{"userName": "yusuf", "userID": 1, "token": "mock-token"},
			wantErr:      false,
		},
		{
			name: "missing templateFile",
			setupFakeDialer: func(f *fakeDialer) {
				f.sent = []*mail.Message{}
			},
			templateFile: "",
			recipient:    "mohamed",
			data:         map[string]any{"userName": "mohamed", "userID": 2, "token": "mock-token"},
			wantErr:      true,
		},
	}

	fake := &fakeDialer{}
	m := &Mailer{dialer: fake, sender: "me@example.com"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFakeDialer(fake)
			gotErr := m.Send(tc.recipient, tc.templateFile, tc.data)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("expecetd error=%v, got %v", tc.wantErr, gotErr)
			}

			// if there was an error, there is no need to perform the next checks
			if gotErr != nil {
				return
			}

			if len(fake.sent) != 1 {
				t.Fatalf("expected 1 email sent, got %d", len(fake.sent))
			}

			msg := fake.sent[0]
			if msg.GetHeader("To")[0] != tc.recipient {
				t.Fatalf("wrong recipient: %v", msg.GetHeader("To")[0])
			}

			expectedSubject := fmt.Sprintf("Hi %s, ", tc.recipient)
			if msg.GetHeader("Subject")[0] != expectedSubject {
				t.Fatalf(
					"expected subject '%s', got '%s'", expectedSubject, msg.GetHeader("Subject")[0],
				)
			}
			buf := new(bytes.Buffer)
			_, err := msg.WriteTo(buf)
			if err != nil {
				t.Errorf("failed recipient write recipient buffer: %v", err)
			}

			if buf.Len() == 0 {
				t.Errorf("body is empty")
			}
		})
	}
}
