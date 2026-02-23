package auth

import "testing"

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		email   string
		wantErr bool
	}{
		{"user@example.com", false},
		{"bad-email", true},
		{"user@", true},
	}
	for _, tc := range cases {
		err := ValidateEmail(tc.email)
		if tc.wantErr && err == nil {
			t.Fatalf("expected error for %q", tc.email)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.email, err)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short", 8); err == nil {
		t.Fatalf("expected weak password error")
	}
	if err := ValidatePassword("longenough", 8); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
