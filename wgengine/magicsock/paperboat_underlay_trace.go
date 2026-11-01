package magicsock

import (
	"bytes"
	"encoding/binary"
	"net/netip"
	"time"

	"tailscale.com/disco"
)

// traceUnderlayPacket retains no packet bodies or keys. Handshake/disco packets are
// complete; encrypted data is sampled to keep a busy transfer's logging bounded.
func (c *Conn) traceUnderlayPacket(inbound bool, b []byte, endpoint netip.AddrPort, failed bool, transport string) {
	if len(b) < 4 {
		return
	}
	kind := binary.LittleEndian.Uint32(b[:4])
	if bytes.HasPrefix(b, []byte(disco.Magic)) {
		kind = 5
	}
	if kind < 1 || kind > 5 {
		return
	}
	direction := "egress"
	last := &c.underlayEgress
	if inbound {
		direction = "ingress"
		last = &c.underlayIngress
	}
	if kind == 4 {
		now := time.Now().UnixNano()
		prior := last.Load()
		if now-prior < int64(100*time.Millisecond) || !last.CompareAndSwap(prior, now) {
			return
		}
	}
	c.logf("paperboat underlay: direction=%s type=%d bytes=%d endpoint=%v failed=%t transport=%s", direction, kind, len(b), endpoint, failed, transport)
}
