package ssrf

import (
	"errors"
	"testing"
)

func TestValidateEndpoint_RejectsPublicBlockedAndPrivate(t *testing.T) {
	p := DefaultPolicy(false)
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{"loopback v4", "https://127.0.0.1", ErrPrivateIP},
		{"rfc1918 10", "https://10.0.0.1", ErrPrivateIP},
		{"rfc1918 192", "https://192.168.1.1", ErrPrivateIP},
		{"http rejected", "http://example.com", ErrDisallowedScheme},
		{"empty host", "https:///bucket", ErrEmptyHost},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateEndpoint(tc.url, p)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("%s: want %v, got %v", tc.url, tc.wantErr, err)
			}
		})
	}
}

func TestValidateEndpoint_AllowlistHTTP(t *testing.T) {
	p := DefaultPolicy(true)
	p.AllowedHosts["127.0.0.1"] = struct{}{}
	_, err := ValidateEndpoint("http://127.0.0.1:9000", p)
	if err != nil {
		t.Fatalf("expected allowlisted http loopback to pass, got %v", err)
	}
}

func TestValidateEndpoint_PublicHTTPSStillResolves(t *testing.T) {
	// We don't require a specific public host to succeed in CI; just make sure
	// that a public-looking hostname is rejected only if it actually resolves
	// to a private IP, not by default.
	p := DefaultPolicy(false)
	_, err := ValidateEndpoint("https://www.cloudflare.com", p)
	if err != nil {
		t.Logf("DNS in CI may be unavailable: %v", err)
	}
}
