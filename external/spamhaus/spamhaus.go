package spamhaus

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type Result struct {
	Success                 bool
	SpamhausBlockList       bool
	CombinedSpamSources     bool
	ExploitsBlockList       bool
	DontRouteOrPeer         bool
	PolicyBlockListISP      bool
	PolicyBlockListSpamhaus bool
	CheckURL                string
}

func (r Result) Summary() string {
	if !r.Success {
		return "check failed"
	}

	var lists []string
	if r.SpamhausBlockList {
		lists = append(lists, "SBL")
	}
	if r.CombinedSpamSources {
		lists = append(lists, "CSS")
	}
	if r.ExploitsBlockList {
		lists = append(lists, "XBL")
	}
	if r.DontRouteOrPeer {
		lists = append(lists, "DROP")
	}
	if r.PolicyBlockListISP {
		lists = append(lists, "PBL")
	}
	if r.PolicyBlockListSpamhaus {
		lists = append(lists, "PBL")
	}

	if len(lists) == 0 {
		return "not listed"
	}

	return "listed: " + strings.Join(lists, ", ")
}

func Check(host string) (Result, error) {
	result := Result{
		CheckURL: fmt.Sprintf("https://check.spamhaus.org/results/?query=%s", host),
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return result, fmt.Errorf("invalid IP address: %s", host)
	}

	var reversed string
	if ip.To4() != nil {
		reversed = reverseIPv4(ip)
	} else {
		reversed = reverseIPv6(ip)
	}

	if reversed == "" {
		return result, fmt.Errorf("failed to reverse IP: %s", host)
	}

	dnsQuery := reversed + ".zen.spamhaus.org"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := &net.Resolver{}
	addrs, err := resolver.LookupHost(ctx, dnsQuery)

	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			result.Success = true
			return result, nil
		}

		return result, fmt.Errorf("spamhaus DNS lookup failed: %w", err)
	}

	for _, addr := range addrs {
		switch addr {
		case "127.0.0.2":
			result.SpamhausBlockList = true
		case "127.0.0.3":
			result.CombinedSpamSources = true
		case "127.0.0.4":
			result.ExploitsBlockList = true
		case "127.0.0.9":
			result.DontRouteOrPeer = true
		case "127.0.0.10":
			result.PolicyBlockListISP = true
		case "127.0.0.11":
			result.PolicyBlockListSpamhaus = true
		}
	}

	result.Success = true
	return result, nil
}

// reverseIPv4 reverses the octets of an IPv4 address for DNSBL lookup
// Example: "204.12.215.98" -> "98.215.12.204"
func reverseIPv4(ip net.IP) string {
	ip = ip.To4()
	if ip == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", ip[3], ip[2], ip[1], ip[0])
}

// reverseIPv6 reverses an IPv6 address for DNSBL lookup
// Example: "2001:db8::1" -> "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2"
func reverseIPv6(ip net.IP) string {
	ip = ip.To16()
	if ip == nil {
		return ""
	}

	var parts []string
	// Convert each byte to two hex digits, then split them
	for i := len(ip) - 1; i >= 0; i-- {
		parts = append(parts, fmt.Sprintf("%x", ip[i]&0x0f))
		parts = append(parts, fmt.Sprintf("%x", ip[i]>>4))
	}

	return strings.Join(parts, ".")
}
