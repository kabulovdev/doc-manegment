package ssrf

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrDisallowedScheme = errors.New("endpoint scheme not allowed (https required)")
	ErrEmptyHost        = errors.New("endpoint host is empty")
	ErrPrivateIP        = errors.New("endpoint resolves to a private or reserved IP")
	ErrNoAddresses      = errors.New("endpoint has no resolvable addresses")
)

type Policy struct {
	AllowHTTP      bool
	AllowedHosts   map[string]struct{} // set of lowercase hostnames for which HTTP is allowed (dev)
	DialTimeout    time.Duration
	KeepAlive      time.Duration
	RequestTimeout time.Duration
}

func DefaultPolicy(allowHTTP bool) Policy {
	return Policy{
		AllowHTTP:      allowHTTP,
		AllowedHosts:   map[string]struct{}{},
		DialTimeout:    10 * time.Second,
		KeepAlive:      30 * time.Second,
		RequestTimeout: 60 * time.Second,
	}
}

func ValidateEndpoint(raw string, p Policy) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return nil, ErrEmptyHost
	}
	if u.Scheme != "https" {
		_, allowed := p.AllowedHosts[host]
		if !p.AllowHTTP && !allowed {
			return nil, ErrDisallowedScheme
		}
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, ErrNoAddresses
	}
	for _, ip := range ips {
		if isDisallowed(ip) {
			if !hostExplicitlyAllowed(host, p) {
				return nil, ErrPrivateIP
			}
		}
	}
	return u, nil
}

func hostExplicitlyAllowed(host string, p Policy) bool {
	_, ok := p.AllowedHosts[host]
	return ok
}

func isDisallowed(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() ||
		ip.IsPrivate() {
		return true
	}
	// CGNAT 100.64.0.0/10
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1]&0xC0 == 64 {
			return true
		}
	}
	return false
}

// GuardedTransport returns an *http.Transport whose DialContext re-verifies the resolved
// IP at connect time, mitigating DNS rebinding between Validate and actual use.
func GuardedTransport(p Policy) *http.Transport {
	base := &net.Dialer{Timeout: p.DialTimeout, KeepAlive: p.KeepAlive}
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, ErrNoAddresses
			}
			for _, ip := range ips {
				if isDisallowed(ip.IP) && !hostExplicitlyAllowed(strings.ToLower(host), p) {
					return nil, ErrPrivateIP
				}
			}
			// Dial the first allowed resolved IP directly (bypasses second DNS lookup).
			return base.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		},
		MaxIdleConns:          50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: p.RequestTimeout,
	}
}
