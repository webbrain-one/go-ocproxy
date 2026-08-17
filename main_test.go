package main

import (
	"strings"
	"testing"
)

func TestResolvedVersionPrefersInjectedValue(t *testing.T) {
	original := version
	version = "test-version"
	t.Cleanup(func() { version = original })
	if got := resolvedVersion(); got != "test-version" {
		t.Fatalf("resolvedVersion = %q, want injected value", got)
	}
}

func TestMTUFromEnv(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fallback int
		value    string
		want     uint32
		wantErr  string
	}{
		{name: "fallback", fallback: 1500, want: 1500},
		{name: "environment", fallback: 1500, value: "1406", want: 1406},
		{name: "invalid", fallback: 1500, value: "bad", wantErr: "invalid INTERNAL_IP4_MTU"},
		{name: "too small", fallback: 575, wantErr: "between 576 and 65535"},
		{name: "too large", fallback: 65536, wantErr: "between 576 and 65535"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := mtuFromEnv(tc.fallback, tc.value)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got (%d, %v), want (%d, nil)", got, err, tc.want)
			}
		})
	}
}
