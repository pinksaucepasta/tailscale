// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package magicsock

import (
	"net"
	"net/netip"
	"reflect"
	"testing"
)

func TestLocalEndpointCandidates(t *testing.T) {
	for _, tt := range []struct {
		name, bind4, bind6  string
		ips, loopback, want []string
		haveEndpoints       bool
	}{
		{name: "different fallback ports", bind4: "0.0.0.0:58442", bind6: "[::]:58441", ips: []string{"192.0.2.1", "2001:db8::1"}, want: []string{"192.0.2.1:58442", "[2001:db8::1]:58441"}},
		{name: "same ports", bind4: "0.0.0.0:41642", bind6: "[::]:41642", ips: []string{"192.0.2.1", "2001:db8::1"}, want: []string{"192.0.2.1:41642", "[2001:db8::1]:41642"}},
		{name: "unavailable ipv4", bind6: "[::]:58441", ips: []string{"192.0.2.1", "2001:db8::1"}, want: []string{"[2001:db8::1]:58441"}},
		{name: "unavailable ipv6", bind4: "0.0.0.0:58442", ips: []string{"192.0.2.1", "2001:db8::1"}, want: []string{"192.0.2.1:58442"}},
		{name: "specific binds", bind4: "192.0.2.2:1234", bind6: "[2001:db8::2]:5678", ips: []string{"192.0.2.1", "2001:db8::1"}, want: []string{"192.0.2.2:1234", "[2001:db8::2]:5678"}},
		{name: "mixed binds", bind4: "192.0.2.2:1234", bind6: "[::]:5678", ips: []string{"192.0.2.1", "2001:db8::1"}, want: []string{"192.0.2.2:1234", "[2001:db8::1]:5678"}},
		{name: "offline loopback", bind4: "0.0.0.0:1234", bind6: "[::]:5678", loopback: []string{"127.0.0.1", "::1"}, want: []string{"127.0.0.1:1234", "[::1]:5678"}},
		{name: "known endpoints suppress loopback", bind4: "0.0.0.0:1234", bind6: "[::]:5678", loopback: []string{"127.0.0.1", "::1"}, haveEndpoints: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bind := func(s string) *net.UDPAddr {
				if s == "" {
					return new(net.UDPAddr)
				}
				return net.UDPAddrFromAddrPort(netip.MustParseAddrPort(s))
			}
			parse := func(ss []string) (out []netip.Addr) {
				for _, s := range ss {
					out = append(out, netip.MustParseAddr(s))
				}
				return
			}
			got, err := localEndpointCandidates(bind(tt.bind4), bind(tt.bind6), tt.haveEndpoints, func() ([]netip.Addr, []netip.Addr, error) { return parse(tt.ips), parse(tt.loopback), nil })
			if err != nil {
				t.Fatal(err)
			}
			var want []netip.AddrPort
			for _, s := range tt.want {
				want = append(want, netip.MustParseAddrPort(s))
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("candidates = %v; want %v", got, want)
			}
		})
	}
}
