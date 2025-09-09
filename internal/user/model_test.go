package user

import "testing"

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		pass    string
		cost    int
		wantErr bool
	}{
		{
			name:    "valid",
			pass:    "12345678",
			cost:    12,
			wantErr: false,
		},
		{
			name:    "cost too low, < 4",
			pass:    "12345678",
			cost:    1,
			wantErr: false, // cost will be set to 10, because cost < 4 isn't allowed
		},
		{
			name:    "cost too high, > 31",
			pass:    "12345678",
			cost:    32,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUser := User{}
			gotErr := mockUser.Password.Set(tc.pass, tc.cost)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tc.wantErr, gotErr)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name          string
		pass          string
		attemptedPass string
		matches       bool
		wantErr       bool
	}{
		{
			name:          "maching password",
			pass:          "12345678",
			attemptedPass: "12345678",
			matches:       true,
		},
		{
			name:          "unmaching password",
			pass:          "12345678",
			attemptedPass: "abcdefhi",
			matches:       false,
		},
	}

	cost := 12
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user := User{}
			user.Password.Set(tc.pass, cost)
			matches, _ := user.Password.Matches(tc.attemptedPass)
			if matches != tc.matches {
				t.Fatalf("expected matches=%v, got matches=%v", tc.matches, matches)
			}
		})
	}
}
