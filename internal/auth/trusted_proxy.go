package auth

import (
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"git.sr.ht/~icikowski/account-center/internal/consts"
)

// TrustedProxies controls whether forwarded headers are accepted for a request.
type TrustedProxies struct {
	networks []*net.IPNet
}

// NewTrustedProxies parses the configured trusted proxy IPs/CIDRs.
func NewTrustedProxies(values []string) (*TrustedProxies, error) {
	if len(values) == 0 {
		// No trusted proxies configured.
		//
		//nolint:nilnil
		return nil, nil
	}

	networks := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		_, network, err := parseTrustedProxyCIDR(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}

	return &TrustedProxies{networks: networks}, nil
}

// AllowsForwardedHeaders checks if the request's remote IP is within any of the trusted proxy networks.
func (t *TrustedProxies) AllowsForwardedHeaders(r *http.Request) bool {
	if t == nil || r == nil {
		return false
	}

	ip := remoteIP(r)
	if ip == nil {
		return false
	}

	return t.contains(ip)
}

// ClientIP resolves the originating client IP for the request.
func (t *TrustedProxies) ClientIP(r *http.Request) net.IP {
	if r == nil {
		return nil
	}

	remote := remoteIP(r)
	if remote == nil {
		return nil
	}

	if !t.AllowsForwardedHeaders(r) {
		return remote
	}

	if ip := t.forwardedForClientIP(r.Header.Get(consts.HeaderXForwardedFor)); ip != nil {
		return ip
	}
	if ip := headerIP(r.Header.Get(consts.HeaderTrueClientIP)); ip != nil {
		return ip
	}
	if ip := headerIP(r.Header.Get(consts.HeaderXRealIP)); ip != nil {
		return ip
	}

	return remote
}

func (t *TrustedProxies) contains(ip net.IP) bool {
	if t == nil || ip == nil {
		return false
	}

	for _, network := range t.networks {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

func remoteIP(r *http.Request) net.IP {
	if r == nil {
		return nil
	}

	host := strings.TrimSpace(r.RemoteAddr)
	if host == "" {
		return nil
	}

	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		host = splitHost
	}

	return net.ParseIP(host)
}

func (t *TrustedProxies) forwardedForClientIP(value string) net.IP {
	if t == nil || value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return nil
	}

	ips := make([]net.IP, 0, len(parts))
	for _, part := range parts {
		ip := headerIP(part)
		if ip == nil {
			return nil
		}
		ips = append(ips, ip)
	}

	for _, v := range slices.Backward(ips) {
		if !t.contains(v) {
			return v
		}
	}

	return ips[0]
}

func headerIP(value string) net.IP {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if ip := net.ParseIP(value); ip != nil {
		return ip
	}
	host, _, err := net.SplitHostPort(value)
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}

func parseTrustedProxyCIDR(value string) (net.IP, *net.IPNet, error) {
	if ip := net.ParseIP(value); ip != nil {
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		return net.ParseCIDR(value + "/" + strconv.Itoa(bits))
	}

	return net.ParseCIDR(value)
}
