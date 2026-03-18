package fetch

import (
	"context"
	"fmt"
	"net"
	"syscall"
)

// safeDialer returns a DialContext that refuses to connect to private, loopback,
// link-local or unspecified addresses. The check runs after DNS resolution on
// the concrete address being dialled, which defends against DNS-rebinding style
// SSRF attacks that a pre-flight hostname lookup would miss.
func safeDialer(allowPrivate bool) func(ctx context.Context, network, addr string) (net.Conn, error) {
	base := &net.Dialer{}
	if allowPrivate {
		return base.DialContext
	}
	base.Control = func(_, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("fetch: could not parse dialled address %q", host)
		}
		if isBlocked(ip) {
			return fmt.Errorf("fetch: refusing to connect to non-public address %s", ip)
		}
		return nil
	}
	return base.DialContext
}

// isBlocked reports whether ip is in a range the fetcher must not reach.
func isBlocked(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		isUniqueLocal(ip)
}

// isUniqueLocal reports whether ip is in the IPv6 unique-local range fc00::/7,
// which net.IP.IsPrivate does cover, but we keep it explicit for clarity.
func isUniqueLocal(ip net.IP) bool {
	if v6 := ip.To16(); v6 != nil && ip.To4() == nil {
		return v6[0]&0xfe == 0xfc
	}
	return false
}
