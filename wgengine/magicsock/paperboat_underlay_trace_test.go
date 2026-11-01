package magicsock

import (
	"encoding/binary"
	"net/netip"
	"testing"

	"tailscale.com/disco"
)

func TestUnderlayTraceRetainsMetadataWithoutPayload(t *testing.T) {
	calls := 0
	c := &Conn{logf: func(format string, args ...any) {
		calls++
		if format != "paperboat underlay: direction=%s type=%d bytes=%d endpoint=%v failed=%t transport=%s" || len(args) != 6 {
			t.Fatal("unexpected trace shape")
		}
		if _, ok := args[0].(string); !ok {
			t.Fatal("missing direction")
		}
		if _, ok := args[1].(uint32); !ok {
			t.Fatal("missing packet type")
		}
		if _, ok := args[2].(int); !ok {
			t.Fatal("missing packet length")
		}
		if _, ok := args[3].(netip.AddrPort); !ok {
			t.Fatal("missing endpoint metadata")
		}
		if _, ok := args[4].(bool); !ok {
			t.Fatal("missing send result")
		}
	}}
	endpoint := netip.MustParseAddrPort("127.0.0.1:1234")
	for kind := uint32(1); kind <= 4; kind++ {
		b := append(make([]byte, 4), []byte("private payload must not become a log argument")...)
		binary.LittleEndian.PutUint32(b, kind)
		c.traceUnderlayPacket(false, b, endpoint, false, "udp")
	}
	c.traceUnderlayPacket(true, append([]byte(disco.Magic), []byte("private disco body")...), endpoint, false, "udp")
	if calls != 5 {
		t.Fatalf("logged %d valid packet metadata records; want5", calls)
	}
	c.traceUnderlayPacket(false, []byte{1}, endpoint, false, "udp")
	c.traceUnderlayPacket(false, []byte("unclassified private body"), endpoint, false, "udp")
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b, 4)
	c.traceUnderlayPacket(false, b, endpoint, false, "udp")
	if calls != 5 {
		t.Fatal("logged invalid packet or bypassed encrypted-data sampling bound")
	}
}
