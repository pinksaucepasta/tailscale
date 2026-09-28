// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package magicsock

import (
	"testing"

	"tailscale.com/health"
	"tailscale.com/net/netcheck"
	"tailscale.com/tailcfg"
	"tailscale.com/tstest"
	"tailscale.com/util/eventbus"
	"tailscale.com/util/eventbus/eventbustest"
)

func TestSTUNOnlyDERPHome(t *testing.T) {
	tstest.Replace(t, &checkControlHealthDuringNearestDERPInTests, true)
	for _, only := range []bool{false, true} {
		t.Run(map[bool]string{false: "mixed", true: "only STUN"}[only], func(t *testing.T) {
			bus := eventbustest.NewBus(t)
			c := newConn(t.Logf)
			c.health = health.NewTracker(bus)
			c.eventBus = bus
			c.eventClient = bus.Client("STUN-only test")
			c.homeDERPChangedPub = eventbus.Publish[HomeDERPChanged](c.eventClient)
			c.derpMap = &tailcfg.DERPMap{Regions: map[tailcfg.DERPRegionID]*tailcfg.DERPRegion{3: {RegionID: 3, Nodes: []*tailcfg.DERPNode{{STUNOnly: true}}}}}
			if !only {
				for _, id := range []tailcfg.DERPRegionID{1, 2} {
					c.derpMap.Regions[id] = &tailcfg.DERPRegion{RegionID: id, Nodes: []*tailcfg.DERPNode{{}}}
				}
			}
			check := func(got tailcfg.DERPRegionID) {
				t.Helper()
				if got == 3 || (only && got != 0) || (!only && got != 1 && got != 2) {
					t.Fatalf("invalid relay home %d", got)
				}
			}
			c.myDerp = 3
			check(c.pickDERPFallback())
			check(c.maybeSetNearestDERP(&netcheck.Report{PreferredDERP: 3}, true))
			check(c.myDerp)
			c.myDerp = 3
			check(c.maybeSetNearestDERP(&netcheck.Report{}, false))
			if c.setNearestDERP(3) || c.myDerp != 0 {
				t.Fatal("direct home selection accepted STUN-only region")
			}
		})
	}
}
