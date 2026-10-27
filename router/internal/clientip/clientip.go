// Package clientip resolves the client address of a request without trusting
// forwarding headers from arbitrary peers.
package clientip

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// Resolver trusts X-Forwarded-For / X-Real-IP only when the direct peer is in
// one of the trusted proxy prefixes.
type Resolver struct {
	trusted []netip.Prefix
}

// NewResolver parses CIDRs or single IPs. Invalid entries cause a panic so
// misconfiguration is caught at startup.
func NewResolver(trustedProxies []string) *Resolver {
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
	return &Resolver{trusted: prefixes}
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

	if xff := r.Header.Values("X-Forwarded-For"); len(xff) > 0 {
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
	if addr, ok := parseAddr(r.Header.Get("X-Real-IP")); ok {
		return addr.String()
	}
	return peer.String()
}
