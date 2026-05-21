package auth

import (
	"net"
	"net/http"
	"strconv"
	"strings"
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
