// Package clientip resolves the client address of a request without trusting
// forwarding headers from arbitrary peers.
package clientip

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// Resolver trusts forwarding headers only when the direct peer is in one of
// the trusted proxy prefixes.
type Resolver struct {
	trusted []netip.Prefix
	// header is the only forwarding header read; empty means X-Forwarded-For
	// with a fallback to X-Real-IP when X-Forwarded-For is absent.
	header string
}

// NewResolver parses CIDRs or single IPs. Invalid entries cause a panic so
// misconfiguration is caught at startup.
//
// header names the single forwarding header the proxy sets ("X-Forwarded-For"
// or "X-Real-IP"). Naming it is recommended: when it is omitted both headers
// are read, so a proxy that sets only X-Real-IP but forwards a client-supplied
// X-Forwarded-For lets clients choose their address.
func NewResolver(trustedProxies []string, header ...string) *Resolver {
	if len(header) > 1 {
		panic("clientip: more than one header provided")
	}
	var forwardedHeader string
	if len(header) == 1 {
		forwardedHeader = http.CanonicalHeaderKey(strings.TrimSpace(header[0]))
		if forwardedHeader != "X-Forwarded-For" && forwardedHeader != "X-Real-Ip" {
			panic("clientip: unsupported forwarding header " + header[0])
		}
	}
	var prefixes []netip.Prefix
	for _, p := range trustedProxies {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.Contains(p, "/") {
			addr, err := netip.ParseAddr(p)
			if err != nil {
				panic("clientip: invalid trusted proxy " + p)
			}
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}
		prefix, err := netip.ParsePrefix(p)
		if err != nil {
			panic("clientip: invalid trusted proxy " + p)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return &Resolver{trusted: prefixes, header: forwardedHeader}
}

func (res *Resolver) isTrusted(addr netip.Addr) bool {
	for _, p := range res.trusted {
		if p.Contains(addr.Unmap()) {
			return true
		}
	}
	return false
}

func parseAddr(s string) (netip.Addr, bool) {
	s = strings.TrimSpace(s)
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

// ClientIP returns the client address. Forwarded headers are walked right to
// left and the first hop that is not a trusted proxy is returned.
func (res *Resolver) ClientIP(r *http.Request) string {
	peer, ok := parseAddr(r.RemoteAddr)
	if !ok {
		return r.RemoteAddr
	}
	if !res.isTrusted(peer) {
		return peer.String()
	}

	if res.header == "X-Real-Ip" {
		if addr, ok := parseAddr(r.Header.Get("X-Real-IP")); ok {
			return addr.String()
		}
		return peer.String()
	}

	xff := r.Header.Values("X-Forwarded-For")
	if len(xff) > 0 {
		hops := strings.Split(strings.Join(xff, ","), ",")
		for i := len(hops) - 1; i >= 0; i-- {
			addr, ok := parseAddr(hops[i])
			if !ok {
				break
			}
			if !res.isTrusted(addr) {
				return addr.String()
			}
		}
	}
	if res.header == "" && len(xff) == 0 {
		if addr, ok := parseAddr(r.Header.Get("X-Real-IP")); ok {
			return addr.String()
		}
	}
	return peer.String()
}
